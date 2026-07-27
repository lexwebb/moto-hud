import java.io.File;
import java.io.FileWriter;
import java.io.PrintWriter;
import java.io.RandomAccessFile;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

import net.osmand.ResultMatcher;
import net.osmand.binary.BinaryMapIndexReader;
import net.osmand.binary.BinaryMapRouteReaderAdapter.RouteRegion;
import net.osmand.binary.BinaryMapRouteReaderAdapter.RouteSubregion;
import net.osmand.binary.RouteDataObject;
import net.osmand.data.LatLon;
import net.osmand.router.RoutePlannerFrontEnd;
import net.osmand.router.RouteSegmentResult;
import net.osmand.router.RoutingConfiguration;
import net.osmand.router.RoutingConfiguration.RoutingMemoryLimits;
import net.osmand.router.RoutingContext;
import net.osmand.router.TurnType;
import net.osmand.util.MapUtils;

/**
 * Off-device OsmAnd-java dump + dual carriageway inference (no dual_carriageway tag).
 */
public class DumpOsmandJunction {

	private static final double PARALLEL_DEG = 18;
	private static final double OPPOSITE_BEARING_MIN = 150;
	private static final double SEP_MIN_M = 10;
	private static final double SEP_MAX_M = 22;
	private static final int APPROACH_LOOKBACK = 6;

	public static void main(String[] args) throws Exception {
		if (args.length < 1) {
			System.err.println("Usage: DumpOsmandJunction <path-to.obf> [out.json]");
			System.exit(2);
		}
		File obf = new File(args[0]);
		File outFile = new File(args.length > 1 ? args[1] : "osmand-junction-dump.json");

		LatLon start = new LatLon(51.50426, -0.126329);
		LatLon end = new LatLon(51.5202, -0.1055);

		RandomAccessFile raf = new RandomAccessFile(obf, "r");
		BinaryMapIndexReader reader = new BinaryMapIndexReader(raf, obf);
		RoutePlannerFrontEnd fe = new RoutePlannerFrontEnd();
		RoutingConfiguration.Builder builder = RoutingConfiguration.getDefault();
		Map<String, String> params = new HashMap<>();
		params.put("car", "true");
		RoutingMemoryLimits mem = new RoutingMemoryLimits(
				RoutingConfiguration.DEFAULT_MEMORY_LIMIT * 4,
				RoutingConfiguration.DEFAULT_NATIVE_MEMORY_LIMIT);
		RoutingConfiguration config = builder.build("car", mem, params);
		RoutingContext ctx = fe.buildRoutingContext(
				config, null, new BinaryMapIndexReader[] { reader },
				RoutePlannerFrontEnd.RouteCalculationMode.NORMAL);
		ctx.leftSideNavigation = true;

		System.out.println("Routing " + start + " → " + end + " on " + obf.getName());
		var calc = fe.searchRoute(ctx, start, end, null);
		if (calc == null || !calc.isCorrect()) {
			System.err.println("Route failed: " + (calc != null ? calc.getError() : "null"));
			System.exit(1);
		}
		List<RouteSegmentResult> segs = calc.getList();
		System.out.println("Segments: " + segs.size());

		List<Map<String, Object>> turns = new ArrayList<>();
		int dualTagHits = 0;
		int dualInferred = 0;
		int withAttachments = 0;

		for (int i = 0; i < segs.size(); i++) {
			RouteSegmentResult seg = segs.get(i);
			TurnType tt = seg.getTurnType();
			if (tt == null && i > 0) {
				continue;
			}
			RouteDataObject obj = seg.getObject();
			String dualTag = obj.getValue("dual_carriageway");
			if ("yes".equals(dualTag)) {
				dualTagHits++;
			}

			List<Map<String, Object>> attached = new ArrayList<>();
			List<RouteSegmentResult> att = seg.getAttachedRoutes(seg.getStartPointIndex());
			if (att != null && !att.isEmpty()) {
				withAttachments++;
			}
			for (RouteSegmentResult a : att) {
				attached.add(armMap(a));
			}

			DualHit dual = inferDual(reader, segs, i);
			if (dual.hit) {
				dualInferred++;
			}

			Map<String, Object> turn = new HashMap<>();
			turn.put("segment_index", i);
			turn.put("turn", tt != null ? tt.toXmlString() : "C");
			turn.put("turn_value", tt != null ? tt.getValue() : TurnType.C);
			turn.put("description", seg.getDescription(false));
			turn.put("highway", obj.getHighway());
			turn.put("oneway", obj.getOneway());
			turn.put("name", safeName(obj));
			turn.put("dual_carriageway_tag", dualTag);
			turn.put("start_lat", seg.getStartPoint().getLatitude());
			turn.put("start_lon", seg.getStartPoint().getLongitude());
			turn.put("bearing_begin", seg.getBearingBegin());
			turn.put("bearing_end", seg.getBearingEnd());
			turn.put("attached_count", attached.size());
			turn.put("attached", attached);
			turn.put("dual_inferred", dual.hit);
			turn.put("dual_reason", dual.reason);
			turns.add(turn);

			System.out.printf(
					"#%d turn=%s hwy=%s oneway=%d name=%s attached=%d dual=%s%n",
					i,
					tt != null ? tt.toXmlString() : "-",
					obj.getHighway(),
					obj.getOneway(),
					safeName(obj),
					attached.size(),
					dual.hit ? dual.reason : "—");
			for (Map<String, Object> am : attached) {
				System.out.println("    + " + am);
			}
		}

		Map<String, Object> report = new HashMap<>();
		report.put("obf", obf.getAbsolutePath());
		report.put("start", Map.of("lat", start.getLatitude(), "lon", start.getLongitude()));
		report.put("end", Map.of("lat", end.getLatitude(), "lon", end.getLongitude()));
		report.put("segment_count", segs.size());
		report.put("turn_count", turns.size());
		report.put("turns_with_attachments", withAttachments);
		report.put("dual_carriageway_tag_hits", dualTagHits);
		report.put("dual_inferred", dualInferred);
		report.put("turns", turns);

		try (PrintWriter pw = new PrintWriter(new FileWriter(outFile))) {
			pw.println(toJson(report));
		}
		System.out.println();
		System.out.println("Wrote " + outFile.getAbsolutePath());
		System.out.println("SUMMARY turns=" + turns.size()
				+ " withAttachments=" + withAttachments
				+ " dualTagHits=" + dualTagHits
				+ " dualInferred=" + dualInferred);
		if (withAttachments == 0) {
			System.err.println("PROOF WEAK: no attached routes found");
			System.exit(1);
		}
		System.out.println("PROOF: OsmAnd-java attachedRoutes OK; dual inference = " + dualInferred);
	}

	/**
	 * Dual inference (no dual_carriageway tag):
	 * 1) turn-local attached/route opposite oneway or opposite-bearing+cross-track
	 * 2) OBF spatial neighborhood (parallel duals often never share nodes)
	 * 3) approach lookback — opposite oneway only (bearing lookback caused FPs)
	 */
	private static DualHit inferDual(BinaryMapIndexReader reader, List<RouteSegmentResult> segs, int turnIdx)
			throws Exception {
		RouteDataObject turnObj = segs.get(turnIdx).getObject();
		// Dual is a property of the corridor we're on, not a side-street turn onto it.
		if (!majorEnough(turnObj.getHighway()) || isLink(turnObj.getHighway())) {
			return DualHit.no();
		}
		DualHit hit = scoreArms(collectArms(segs, turnIdx, 0), turnIdx, true);
		if (hit.hit) {
			return hit;
		}
		hit = inferDualSpatial(reader, segs.get(turnIdx));
		if (hit.hit) {
			return hit;
		}
		return scoreArms(collectArms(segs, turnIdx, APPROACH_LOOKBACK), turnIdx, false);
	}

	/** Load nearby RouteDataObjects and look for opposite parallel major/same-name. */
	private static DualHit inferDualSpatial(BinaryMapIndexReader reader, RouteSegmentResult seg) throws Exception {
		RouteDataObject selfObj = seg.getObject();
		Arm self = armOf(seg, "route", -1);
		if (!majorEnough(self.highway) || isLink(self.highway)) {
			return DualHit.no();
		}
		LatLon p = seg.getStartPoint();
		List<RouteDataObject> nearby = loadNearbyRoutes(reader, p.getLatitude(), p.getLongitude(), 50);
		long selfId = selfObj.getId();
		for (RouteDataObject other : nearby) {
			if (other.getId() == selfId) {
				continue;
			}
			if (!isCarHighway(other.getHighway()) || isLink(other.getHighway())) {
				continue;
			}
			Arm o = armFromRdo(other, "spatial");
			if (o == null) {
				continue;
			}
			// On a bidirectional carriageway, ignore oneway slips/cycle-adjacent primaries.
			// (Farringdon FP: primary oneway=1 parallel to undivided primary.)
			// When we are already oneway, allow opposite with missing oneway tag (Whitehall OBF).
			if (self.oneway == 0 && o.oneway != 0) {
				continue;
			}
			// Undivided continuity: shared nodes + opposite bearing ≠ dual
			if (!(self.oneway != 0 && o.oneway != 0 && oppositeBearing(self.bearing, o.bearing))
					&& shareAnyNode(selfObj, other)) {
				continue;
			}
			DualHit cand = scorePair(self, o, -1, true);
			if (cand.hit) {
				return DualHit.yes("spatial:" + cand.reason);
			}
		}
		return DualHit.no();
	}

	private static boolean shareAnyNode(RouteDataObject a, RouteDataObject b) {
		for (int i = 0; i < a.getPointsLength(); i++) {
			int ax = a.getPoint31XTile(i);
			int ay = a.getPoint31YTile(i);
			for (int j = 0; j < b.getPointsLength(); j++) {
				if (ax == b.getPoint31XTile(j) && ay == b.getPoint31YTile(j)) {
					return true;
				}
			}
		}
		return false;
	}

	private static List<RouteDataObject> loadNearbyRoutes(
			BinaryMapIndexReader reader, double lat, double lon, double radiusM) throws Exception {
		double dLat = radiusM / 111320.0;
		double dLon = radiusM / (111320.0 * Math.cos(Math.toRadians(lat)));
		int left = MapUtils.get31TileNumberX(lon - dLon);
		int right = MapUtils.get31TileNumberX(lon + dLon);
		int top = MapUtils.get31TileNumberY(lat + dLat);
		int bottom = MapUtils.get31TileNumberY(lat - dLat);
		final List<RouteDataObject> out = new ArrayList<>();
		ResultMatcher<RouteDataObject> matcher = new ResultMatcher<RouteDataObject>() {
			@Override
			public boolean publish(RouteDataObject object) {
				out.add(object);
				return false;
			}

			@Override
			public boolean isCancelled() {
				return false;
			}
		};
		BinaryMapIndexReader.SearchRequest<RouteDataObject> req =
				BinaryMapIndexReader.buildSearchRouteRequest(left, right, top, bottom, matcher);
		List<RouteSubregion> toLoad = new ArrayList<>();
		for (RouteRegion reg : reader.getRoutingIndexes()) {
			List<RouteSubregion> roots = new ArrayList<>(reg.getSubregions());
			toLoad.addAll(reader.searchRouteIndexTree(req, roots));
		}
		reader.loadRouteIndexData(toLoad, matcher);
		return out;
	}

	private static Arm armFromRdo(RouteDataObject o, String src) {
		if (o.getPointsLength() < 2) {
			return null;
		}
		double lat0 = MapUtils.get31LatitudeY(o.getPoint31YTile(0));
		double lon0 = MapUtils.get31LongitudeX(o.getPoint31XTile(0));
		int i1 = Math.min(1, o.getPointsLength() - 1);
		double lat1 = MapUtils.get31LatitudeY(o.getPoint31YTile(i1));
		double lon1 = MapUtils.get31LongitudeX(o.getPoint31XTile(i1));
		float brg = (float) bearingBetween(lat0, lon0, lat1, lon1);
		if (o.getOneway() < 0) {
			brg = (float) bearingBetween(lat1, lon1, lat0, lon0);
		}
		// Local midpoint (first ~40 m of way), not whole-way centroid
		int iMid = Math.min(Math.max(1, o.getPointsLength() / 8), o.getPointsLength() - 1);
		double latM = MapUtils.get31LatitudeY(o.getPoint31YTile(iMid));
		double lonM = MapUtils.get31LongitudeX(o.getPoint31XTile(iMid));
		return new Arm(src, -1, safeName(o), o.getHighway(), o.getOneway(), brg,
				lat0, lon0, latM, lonM);
	}

	private static List<Arm> collectArms(List<RouteSegmentResult> segs, int turnIdx, int lookback) {
		List<Arm> arms = new ArrayList<>();
		int from = Math.max(0, turnIdx - lookback);
		for (int i = from; i <= turnIdx; i++) {
			RouteSegmentResult s = segs.get(i);
			arms.add(armOf(s, "route:" + i, i));
			int start = Math.min(s.getStartPointIndex(), s.getEndPointIndex());
			int end = Math.max(s.getStartPointIndex(), s.getEndPointIndex());
			for (int pi = start; pi <= end; pi++) {
				addAttached(arms, s, pi, "att:" + i + "@" + pi, i);
			}
		}
		return arms;
	}

	private static DualHit scoreArms(List<Arm> arms, int turnIdx, boolean allowBearingSep) {
		DualHit best = DualHit.no();
		for (int a = 0; a < arms.size(); a++) {
			for (int b = a + 1; b < arms.size(); b++) {
				Arm x = arms.get(a);
				Arm y = arms.get(b);
				// Lookback must not stick a prior dual onto a later turn (Northumberland→Embankment).
				if (!allowBearingSep && x.segIdx != turnIdx && y.segIdx != turnIdx) {
					continue;
				}
				DualHit cand = scorePair(x, y, turnIdx, allowBearingSep);
				if (!cand.hit) {
					continue;
				}
				boolean candAtTurn = x.segIdx == turnIdx || y.segIdx == turnIdx;
				boolean bestAtTurn = best.hit && best.reason != null && best.reason.contains("@turn");
				if (!best.hit || (candAtTurn && !bestAtTurn)) {
					best = candAtTurn ? DualHit.yes(cand.reason + "@turn") : cand;
				}
			}
		}
		return best;
	}

	private static DualHit scorePair(Arm x, Arm y, int turnIdx, boolean allowBearingSep) {
		if (!majorEnough(x.highway) && !majorEnough(y.highway)) {
			return DualHit.no();
		}
		if (isLink(x.highway) != isLink(y.highway) && sameName(x.name, y.name)) {
			if (!oppositeOneway(x, y)) {
				return DualHit.no();
			}
		}
		boolean named = sameName(x.name, y.name);
		boolean oppOne = oppositeOneway(x, y);
		boolean oppBrg = oppositeBearing(x.bearing, y.bearing);
		if (!oppOne && !oppBrg) {
			return DualHit.no();
		}
		if (!parallelOrAnti(x.bearing, y.bearing)) {
			return DualHit.no();
		}

		double latSep = lateralSeparationM(x, y);
		double nodeDist = MapUtils.getDistance(x.lat, x.lon, y.lat, y.lon);

		if (oppOne && (named || (majorEnough(x.highway) && majorEnough(y.highway)))) {
			if (nodeDist <= SEP_MAX_M || (latSep >= SEP_MIN_M && latSep <= SEP_MAX_M)) {
				String label = named ? x.name : "major";
				return DualHit.yes("opposite_oneway:" + label);
			}
		}

		if (allowBearingSep && oppBrg && latSep >= SEP_MIN_M && latSep <= SEP_MAX_M) {
			// Both-bidirectional parallels are too weak (Farringdon FP). Need a oneway signal.
			if (x.oneway == 0 && y.oneway == 0) {
				return DualHit.no();
			}
			if (named) {
				return DualHit.yes("opposite_bearing_sep:" + x.name);
			}
			if (majorEnough(x.highway) && majorEnough(y.highway)) {
				return DualHit.yes("major_opposite_bearing_sep");
			}
		}
		return DualHit.no();
	}

	private static void addAttached(List<Arm> arms, RouteSegmentResult host, int idx, String src, int segIdx) {
		List<RouteSegmentResult> att = host.getAttachedRoutes(idx);
		if (att == null) {
			return;
		}
		for (RouteSegmentResult a : att) {
			arms.add(armOf(a, src, segIdx));
		}
	}

	private static Arm armOf(RouteSegmentResult s, String src, int segIdx) {
		RouteDataObject o = s.getObject();
		LatLon start = s.getStartPoint();
		LatLon end = s.getEndPoint();
		double midLat = (start.getLatitude() + end.getLatitude()) / 2.0;
		double midLon = (start.getLongitude() + end.getLongitude()) / 2.0;
		return new Arm(
				src,
				segIdx,
				safeName(o),
				o.getHighway(),
				o.getOneway(),
				s.getBearingBegin(),
				start.getLatitude(),
				start.getLongitude(),
				midLat,
				midLon);
	}

	/** Cross-track only — never use along-track midpoint distance. */
	private static double lateralSeparationM(Arm a, Arm b) {
		double ctA = crossTrackM(a.lat, a.lon, a.bearing, b.midLat, b.midLon);
		double ctB = crossTrackM(b.lat, b.lon, b.bearing, a.midLat, a.midLon);
		return (ctA + ctB) / 2.0;
	}

	/** Absolute cross-track distance from a point to a ray (lat,lon)+bearing. */
	private static double crossTrackM(double lat0, double lon0, float bearingDeg, double lat, double lon) {
		double dist = MapUtils.getDistance(lat0, lon0, lat, lon);
		if (dist < 1) {
			return 0;
		}
		double brgTo = bearingBetween(lat0, lon0, lat, lon);
		double delta = bearingDelta(bearingDeg, (float) brgTo);
		return Math.abs(Math.sin(Math.toRadians(delta)) * dist);
	}

	/** Initial bearing degrees from (lat1,lon1) to (lat2,lon2). */
	private static double bearingBetween(double lat1, double lon1, double lat2, double lon2) {
		double φ1 = Math.toRadians(lat1);
		double φ2 = Math.toRadians(lat2);
		double Δλ = Math.toRadians(lon2 - lon1);
		double y = Math.sin(Δλ) * Math.cos(φ2);
		double x = Math.cos(φ1) * Math.sin(φ2) - Math.sin(φ1) * Math.cos(φ2) * Math.cos(Δλ);
		return Math.toDegrees(Math.atan2(y, x));
	}

	private static boolean sameName(String a, String b) {
		if (a == null || b == null) {
			return false;
		}
		return a.equalsIgnoreCase(b);
	}

	private static boolean isLink(String hwy) {
		return hwy != null && hwy.endsWith("_link");
	}

	private static boolean majorEnough(String hwy) {
		if (hwy == null) {
			return false;
		}
		return hwy.equals("motorway") || hwy.equals("motorway_link")
				|| hwy.equals("trunk") || hwy.equals("trunk_link")
				|| hwy.equals("primary") || hwy.equals("primary_link")
				|| hwy.equals("secondary") || hwy.equals("secondary_link");
	}

	private static boolean isCarHighway(String hwy) {
		return majorEnough(hwy)
				|| "tertiary".equals(hwy) || "tertiary_link".equals(hwy)
				|| "unclassified".equals(hwy) || "residential".equals(hwy);
	}

	/**
	 * Both ways are oneway and travel in opposite directions.
	 * Note: each dual carriageway is typically oneway=+1 along its own geometry;
	 * opposite travel is detected via bearings, not opposite oneway signs.
	 */
	private static boolean oppositeOneway(Arm a, Arm b) {
		if (a.oneway == 0 || b.oneway == 0) {
			return false;
		}
		return oppositeBearing(a.bearing, b.bearing);
	}

	private static boolean oppositeBearing(float a, float b) {
		return Math.abs(bearingDelta(a, b)) >= OPPOSITE_BEARING_MIN;
	}

	private static boolean parallelOrAnti(float a, float b) {
		double d = Math.abs(bearingDelta(a, b));
		return d <= PARALLEL_DEG || d >= (180 - PARALLEL_DEG);
	}

	private static double bearingDelta(float a, float b) {
		double d = a - b;
		while (d > 180) {
			d -= 360;
		}
		while (d < -180) {
			d += 360;
		}
		return d;
	}

	private static Map<String, Object> armMap(RouteSegmentResult a) {
		RouteDataObject ao = a.getObject();
		Map<String, Object> am = new HashMap<>();
		am.put("highway", ao.getHighway());
		am.put("oneway", ao.getOneway());
		am.put("name", safeName(ao));
		am.put("dual_carriageway", ao.getValue("dual_carriageway"));
		am.put("bearing_begin", a.getBearingBegin());
		return am;
	}

	private static String safeName(RouteDataObject obj) {
		try {
			String n = obj.getName();
			return n == null || n.isEmpty() ? null : n;
		} catch (Throwable t) {
			return null;
		}
	}

	private static final class Arm {
		final String src;
		final int segIdx;
		final String name;
		final String highway;
		final int oneway;
		final float bearing;
		final double lat;
		final double lon;
		final double midLat;
		final double midLon;

		Arm(String src, int segIdx, String name, String highway, int oneway, float bearing,
				double lat, double lon, double midLat, double midLon) {
			this.src = src;
			this.segIdx = segIdx;
			this.name = name;
			this.highway = highway;
			this.oneway = oneway;
			this.bearing = bearing;
			this.lat = lat;
			this.lon = lon;
			this.midLat = midLat;
			this.midLon = midLon;
		}
	}

	private static final class DualHit {
		final boolean hit;
		final String reason;

		DualHit(boolean hit, String reason) {
			this.hit = hit;
			this.reason = reason;
		}

		static DualHit yes(String reason) {
			return new DualHit(true, reason);
		}

		static DualHit no() {
			return new DualHit(false, null);
		}
	}

	@SuppressWarnings("unchecked")
	private static String toJson(Object o) {
		if (o == null) {
			return "null";
		}
		if (o instanceof String) {
			return "\"" + ((String) o).replace("\\", "\\\\").replace("\"", "\\\"") + "\"";
		}
		if (o instanceof Number || o instanceof Boolean) {
			return o.toString();
		}
		if (o instanceof List) {
			StringBuilder sb = new StringBuilder("[");
			boolean first = true;
			for (Object x : (List<?>) o) {
				if (!first) {
					sb.append(",");
				}
				first = false;
				sb.append(toJson(x));
			}
			return sb.append("]").toString();
		}
		if (o instanceof Map) {
			StringBuilder sb = new StringBuilder("{");
			boolean first = true;
			for (Map.Entry<?, ?> e : ((Map<?, ?>) o).entrySet()) {
				if (!first) {
					sb.append(",");
				}
				first = false;
				sb.append(toJson(String.valueOf(e.getKey()))).append(":").append(toJson(e.getValue()));
			}
			return sb.append("}").toString();
		}
		return toJson(String.valueOf(o));
	}
}
