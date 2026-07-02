import esbuild from "esbuild"

const stubUnresolved = {
  name: "stub-unresolved",
  setup(build) {
    build.onResolve({ filter: /^[^.]/ }, async (args) => {
      if (args.path.startsWith("node:")) return null
      if (args.pluginData?.stubProbe) return null
      const r = await build.resolve(args.path, {
        kind: args.kind,
        resolveDir: args.resolveDir,
        importer: args.importer,
        pluginData: { stubProbe: true },
      })
      if (r.errors.length > 0) {
        console.log("[stub] unresolved -> empty module:", args.path)
        return { path: args.path, namespace: "stub" }
      }
      return r
    })
    build.onLoad({ filter: /.*/, namespace: "stub" }, () => ({
      contents: "module.exports = new Proxy(function(){}, { get: () => (function(){}) });",
      loader: "js",
    }))
  },
}

await esbuild.build({
  entryPoints: ["src/index.ts"],
  bundle: true,
  platform: "node",
  format: "esm",
  target: "node20",
  outfile: "bundle.new.mjs",
  plugins: [stubUnresolved],
  logLevel: "warning",
  banner: {
    js: [
      "import { createRequire as _createRequire } from 'module';",
      "import { fileURLToPath as _fileURLToPath } from 'url';",
      "import { dirname as _dirname } from 'path';",
      "const require = _createRequire(import.meta.url);",
      "const __filename = _fileURLToPath(import.meta.url);",
      "const __dirname = _dirname(__filename);",
    ].join("\n"),
  },
})

console.log("BUILD_OK")
