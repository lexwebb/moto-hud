import { useEffect, useRef, useState } from 'react';
import * as THREE from 'three';
import { OrbitControls } from 'three/examples/jsm/controls/OrbitControls.js';
import { GLTFLoader } from 'three/examples/jsm/loaders/GLTFLoader.js';
import { STLLoader } from 'three/examples/jsm/loaders/STLLoader.js';

const MODELS = [
  { id: 'bench', label: 'Bench', file: 'assembly.glb' },
  { id: 'dock', label: 'Dock slabs', file: 'dock_interface.stl' },
  { id: 'plate', label: 'Bike plate', file: 'dock_plate_slab.stl' },
  { id: 'pod', label: 'Pod face', file: 'dock_pod_slab.stl' },
] as const;

type ModelId = (typeof MODELS)[number]['id'];

export default function EnclosureViewer() {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const statusRef = useRef<HTMLDivElement>(null);
  const [modelId, setModelId] = useState<ModelId>('dock');

  useEffect(() => {
    const canvas = canvasRef.current;
    const status = statusRef.current;
    if (!canvas) return;

    const renderer = new THREE.WebGLRenderer({ canvas, antialias: true });
    renderer.setPixelRatio(Math.min(window.devicePixelRatio, 2));
    renderer.setClearColor(0x12151a, 1);
    renderer.shadowMap.enabled = true;
    renderer.shadowMap.type = THREE.PCFSoftShadowMap;
    renderer.toneMapping = THREE.ACESFilmicToneMapping;
    renderer.toneMappingExposure = 1.05;
    renderer.outputColorSpace = THREE.SRGBColorSpace;

    const scene = new THREE.Scene();
    scene.background = new THREE.Color(0x12151a);

    const camera = new THREE.PerspectiveCamera(40, 1, 0.001, 20);
    camera.position.set(0.11, 0.08, 0.13);

    const controls = new OrbitControls(camera, canvas);
    controls.enableDamping = true;
    controls.dampingFactor = 0.08;
    controls.minDistance = 0.05;
    controls.maxDistance = 0.6;

    scene.add(new THREE.AmbientLight(0xffffff, 0.15));

    const key = new THREE.DirectionalLight(0xffffff, 2.6);
    key.castShadow = true;
    key.shadow.mapSize.set(2048, 2048);
    key.shadow.bias = -0.0001;
    key.shadow.normalBias = 0.001;
    scene.add(key);
    scene.add(key.target);

    const fill = new THREE.DirectionalLight(0x9eb6d4, 0.35);
    fill.position.set(-1, 0.4, -0.5);
    scene.add(fill);

    const ground = new THREE.Mesh(
      new THREE.CircleGeometry(0.4, 48),
      new THREE.ShadowMaterial({ opacity: 0.5 }),
    );
    ground.rotation.x = -Math.PI / 2;
    ground.receiveShadow = true;
    scene.add(ground);

    const plastic = new THREE.MeshStandardMaterial({
      color: 0x9aa3ae,
      metalness: 0.05,
      roughness: 0.65,
      flatShading: true,
    });

    let loaded: THREE.Object3D | null = null;
    let raf = 0;
    let alive = true;

    function placeKeyLight(object: THREE.Object3D) {
      const box = new THREE.Box3().setFromObject(object);
      const size = box.getSize(new THREE.Vector3());
      const center = box.getCenter(new THREE.Vector3());
      const extent = Math.max(size.x, size.y, size.z);

      key.target.position.copy(center);
      key.position.set(
        center.x + extent * 1.2,
        center.y + extent * 2.2,
        center.z + extent * 1.4,
      );
      key.target.updateMatrixWorld();

      const cam = key.shadow.camera as THREE.OrthographicCamera;
      const span = extent * 1.4;
      cam.near = extent * 0.1;
      cam.far = extent * 6;
      cam.left = -span;
      cam.right = span;
      cam.top = span;
      cam.bottom = -span;
      cam.updateProjectionMatrix();
      key.shadow.needsUpdate = true;
    }

    function fitObject(root: THREE.Object3D) {
      const box0 = new THREE.Box3().setFromObject(root);
      const size0 = box0.getSize(new THREE.Vector3());
      const scale = 0.12 / Math.max(size0.x, size0.y, size0.z, 1e-6);
      root.scale.setScalar(scale);
      root.updateMatrixWorld(true);

      const box1 = new THREE.Box3().setFromObject(root);
      const center1 = box1.getCenter(new THREE.Vector3());
      root.position.x -= center1.x;
      root.position.z -= center1.z;
      root.position.y -= box1.min.y;
      root.updateMatrixWorld(true);

      root.traverse((obj) => {
        if (!(obj as THREE.Mesh).isMesh) return;
        const mesh = obj as THREE.Mesh;
        mesh.castShadow = true;
        mesh.receiveShadow = true;
        mesh.material = plastic.clone();
        mesh.geometry.computeVertexNormals();
      });

      scene.add(root);
      loaded = root;
      placeKeyLight(root);
      const h = new THREE.Box3().setFromObject(root).getSize(new THREE.Vector3()).y;
      controls.target.set(0, h * 0.4, 0);
      controls.update();
    }

    function clearLoaded() {
      if (!loaded) return;
      scene.remove(loaded);
      loaded.traverse((obj) => {
        const mesh = obj as THREE.Mesh;
        if (mesh.isMesh) {
          mesh.geometry.dispose();
          const mat = mesh.material;
          if (Array.isArray(mat)) mat.forEach((m) => m.dispose());
          else mat.dispose();
        }
      });
      loaded = null;
    }

    function resize() {
      const parent = canvas.parentElement;
      const w = parent?.clientWidth || canvas.clientWidth;
      const h = parent?.clientHeight || canvas.clientHeight;
      renderer.setSize(w, h, false);
      camera.aspect = w / Math.max(h, 1);
      camera.updateProjectionMatrix();
    }

    window.addEventListener('resize', resize);
    const ro = new ResizeObserver(resize);
    if (canvas.parentElement) ro.observe(canvas.parentElement);
    resize();

    const spec = MODELS.find((m) => m.id === modelId) ?? MODELS[0];
    const base = import.meta.env.BASE_URL;
    const url = `${base}enclosure/${spec.file}`;
    if (status) status.textContent = 'Loading…';

    const onError = (err: unknown) => {
      console.error(err);
      if (status) status.textContent = `Failed to load ${spec.file}`;
    };

    if (spec.file.endsWith('.stl')) {
      new STLLoader().load(
        url,
        (geom) => {
          if (!alive) {
            geom.dispose();
            return;
          }
          clearLoaded();
          const mesh = new THREE.Mesh(geom, plastic.clone());
          fitObject(mesh);
          if (status) status.textContent = '';
        },
        undefined,
        onError,
      );
    } else {
      new GLTFLoader().load(
        url,
        (gltf) => {
          if (!alive) return;
          clearLoaded();
          fitObject(gltf.scene);
          if (status) status.textContent = '';
        },
        undefined,
        onError,
      );
    }

    const frame = () => {
      controls.update();
      renderer.render(scene, camera);
      raf = requestAnimationFrame(frame);
    };
    frame();

    return () => {
      alive = false;
      cancelAnimationFrame(raf);
      window.removeEventListener('resize', resize);
      ro.disconnect();
      clearLoaded();
      controls.dispose();
      renderer.dispose();
      plastic.dispose();
      ground.geometry.dispose();
      (ground.material as THREE.Material).dispose();
    };
  }, [modelId]);

  return (
    <div className="viewer-wrap">
      <div className="model-switch" role="tablist" aria-label="Enclosure model">
        {MODELS.map((m) => (
          <button
            key={m.id}
            type="button"
            role="tab"
            aria-selected={modelId === m.id}
            className={modelId === m.id ? 'on' : undefined}
            onClick={() => setModelId(m.id)}
          >
            {m.label}
          </button>
        ))}
      </div>
      <canvas ref={canvasRef} id="viewport" />
      <div ref={statusRef} className="viewer-status">
        Loading…
      </div>
      <style>{`
        .viewer-wrap {
          position: relative;
          flex: 1;
          min-height: 320px;
        }
        .viewer-wrap canvas {
          display: block;
          width: 100%;
          height: 100%;
        }
        .model-switch {
          position: absolute;
          top: 0.75rem;
          left: 0.75rem;
          z-index: 2;
          display: flex;
          flex-wrap: wrap;
          gap: 0.35rem;
        }
        .model-switch button {
          background: #1c1f24;
          color: #e8eaed;
          border: 1px solid #2c3036;
          border-radius: 999px;
          padding: 0.28rem 0.7rem;
          font: inherit;
          font-size: 0.8rem;
          cursor: pointer;
        }
        .model-switch button.on {
          background: #8ab4f8;
          color: #12151a;
          border-color: #8ab4f8;
        }
        .viewer-status {
          position: absolute;
          left: 1rem;
          bottom: 1rem;
          color: #9aa0a6;
          font-size: 0.8rem;
          pointer-events: none;
        }
      `}</style>
    </div>
  );
}
