package partials

templ BuildResults() {
<section
  id="build-results"
  hx-on::after-settle="document.getElementById('build-results').scrollIntoView({ behavior: 'smooth' });"
  class="container mx-auto p-4"
>
  <h3 class="text-4xl font-bold mb-4">
    Bundles to fit: {{ .RequestedBundleSize }}cm
  </h3>

  <div id="build-results-content" class="md:flex gap-4">
    {{ range .Bundles }}
    <div
      style="animation: fadeIn; border: 1px solid gray"
      class="bg-white rounded-lg p-4 mb-8"
    >
      <h2 class="text-2xl font-bold mb-4">{{ .Name }} {{ title .Color }}</h2>
      <ul>
        <li>Total Bundle Price €{{ .Price }}</li>
        <li>Width: {{ sizeRange .Width .Tolerance }} - {{ .Width }}cm</li>
      </ul>
      <strong class="py-4 font-medium mt-4 block">Bundle Includes:</strong>

      <div class="flex flex-col">
        {{ range .Components }}
        {{ template "bundle-component-card" . }}
        {{ end }}
      </div>

      <form
        hx-trigger="submit"
        hx-target="#cart-modal"
        hx-swap="outerHTML"
        hx-post="/cart/add"
        class="flex justify-end"
      >
 
      {{ range .Components }}
          <input type="hidden" name="data" value='{{ toString .}}' />
      {{ end }}


        <button
          class="hover:bg-gray-700 text-white font-bold py-2 px-4 rounded"
          style="background-color: #683b1c"
        >
          Add Bundle To Cart
        </button>
      </form>
    </div>
    {{ end }}
  </div>
</section>

  }

