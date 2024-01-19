import { defineNuxtModule, createResolver, isNuxt2, resolveModule, addPlugin, addImports, addImportsDir } from '@nuxt/kit';

const module = defineNuxtModule({
  meta: {
    name: "pinia",
    configKey: "pinia",
    compatibility: {
      nuxt: "^2.0.0 || ^3.0.0-rc.5",
      bridge: true
    }
  },
  defaults: {
    disableVuex: true
  },
  setup(options, nuxt) {
    const resolver = createResolver(import.meta.url);
    if (
      // @ts-expect-error: no feature flag anymore or private?
      nuxt.options.features && // ts
      options.disableVuex && isNuxt2()
    ) {
      nuxt.options.features.store = false;
    }
    nuxt.options.build.transpile.push(resolver.resolve("./runtime"));
    nuxt.options.alias.pinia = nuxt.options.alias.pinia || // FIXME: remove this deprecated call. Ensure it works in Nuxt 2 to 3
    resolveModule("pinia/dist/pinia.mjs", {
      paths: [nuxt.options.rootDir, import.meta.url]
    });
    nuxt.hook("prepare:types", ({ references }) => {
      references.push({ types: "@pinia/nuxt" });
    });
    nuxt.hook("modules:done", () => {
      if (isNuxt2()) {
        addPlugin(resolver.resolve("./runtime/plugin.vue2"));
      } else {
        addPlugin(resolver.resolve("./runtime/plugin.vue3"));
      }
    });
    const composables = resolver.resolve("./runtime/composables");
    addImports([
      { from: composables, name: "defineStore" },
      { from: composables, name: "acceptHMRUpdate" },
      { from: composables, name: "usePinia" },
      { from: composables, name: "storeToRefs" }
    ]);
    if (!options.storesDirs) {
      options.storesDirs = [resolver.resolve(nuxt.options.srcDir, "stores")];
    }
    if (options.storesDirs) {
      for (const storeDir of options.storesDirs) {
        addImportsDir(resolver.resolve(nuxt.options.rootDir, storeDir));
      }
    }
  }
});

export { module as default };
