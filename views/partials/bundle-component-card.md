
package partials

templ BundleComponentCard() {
  <div class="px-2 mb-4 relative inline-block">
    <div class="flex justify-between items-centerbg-white rounded-lg overflow-hidden shadow-md">
        <a class="relative" href="/{{ .Type }}s/{{ .Id }}">
            <img class="h-full aspect-square max-w-16"
             src="https://cdn11.bigcommerce.com/s-egiahb/images/stencil/640w/products/347/1003/baby-dan-premier-60114-white-no-child__43601.1559815258.jpg?c=2" 
             alt="Baby Safety Gate" class="w-full">
        </a>
        <div class="p-4">
            <a href="/gates/{{.Id}}">
                <h3 class="font-bold mb-2 text-xs md:text-base">
                    {{ .Name }} {{ title  .Color }}
                </h3>
            </a>
        </div>
        <span 
        style="background-color: #683B1C;"
        class=" p-4 z-10 text-white rounded text-xs md:text-base text-nowrap"> x {{ .Qty }}</span>
    </div>
</div>

}

