interface Product {
  id: number,
  name: string,
  width: number,
  price: number,
  img: string,
  color: string,
  qty: number
};

interface Gate extends Product {
  tolerance: number
};

type Gates = Gate[];

interface Extension extends Product {
};


type Extensions = Extension[];

interface Bundle extends Product {
  gates: Gates,
  extensions: Extensions,
  max_width: number
};

type Bundles = Bundle[];


const buildForm = document.querySelector("#build-gate");
if (buildForm) {
  buildForm.addEventListener("submit", async function (e) {
    e.preventDefault();
    try {
      const input = buildForm.querySelector(
        "#desired-width"
      ) as HTMLInputElement;
      const width = parseInt(input?.value);

      const res = await fetch("/build/", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ width }),
      });

      if (res.ok) {
        const data = (await res.json()) as Bundles;

        const elems = [];
        for (let i = 0; i < data.length; i++) {
          const bundle = data[i];

          const params = new URLSearchParams();

          const extensions = [];
          for (let ii = 0; ii < bundle.extensions.length; ii++) {
            const extension = bundle.extensions[ii];
            extensions.push({
              id: extension.id,
              qty: extension.qty,
            });
          }

          params.set(
            "gate",
            JSON.stringify({ id: bundle.gates[0].id, qty: bundle.gates[0].qty })
          );
          params.set("extensions", JSON.stringify(extensions));

          const html = /*htm*/ `
          <a id="fade-in-element" class="w-full md:w-1/2 lg:w-1/4 px-2 mb-4 hidden" href="/bundles/?${params.toString()}" >
            <div class="bg-white rounded-lg overflow-hidden shadow-md">
                <img src="https://via.placeholder.com/500x300" alt="Baby Safety Gate" class="w-full">
                <div class="p-4">
                    <h3 class="font-bold mb-2">${bundle.gates[0].name} ${
            bundle.gates[0].color
          } ${
            bundle.extensions.length
              ? " &amp; " +
                bundle.extensions.reduce((a, c) => a + c.qty, 0) +
                " Extensions"
              : ""
          }</h3>
                    <p class="text-gray-600 mb-4">This baby safety gate is perfect for keeping your baby safe in any room of your house.</p>
                    <div class="flex justify-between items-center">
                        <span class="text-xl font-bold">€${bundle.price}</span>
                        <button class="atc-btn bg-gray-800 hover:bg-gray-700 text-white font-bold py-2 px-4 rounded">Add to Cart</button>
                    </div>
                </div>
            </div>
          </a>`;
          elems.push(html);
        }
        const htmlToInsert = elems.join("");

        const resultsArea = document.querySelector("#build-results");
        if (resultsArea) {
          resultsArea.classList.remove("hidden");

          const container = document.querySelector("#build-results-content");
          if (container) {
            container.innerHTML = htmlToInsert;

            const fadeElems = container.querySelectorAll("#fade-in-element");
            if (fadeElems) {
              for (let x = 0; x < fadeElems?.length; x++) {
                const fElem = fadeElems[x];
                await new Promise(function (resolve) {
                  fElem.classList.remove("hidden");
                  fElem.addEventListener("animationend", function () {
                    fElem.classList.remove("fade-in");
                    resolve(null);
                  });
                  fElem.classList.add("fade-in");
                });
              }
            }
          }
        }
      }
    } catch (err) {
      console.log(err);
    }
  });
} else {
  console.log("no build form");
}

/**add to cart functionality */

try {
  const cart = JSON.parse(localStorage.getItem("cart") || "[]");

  const addToCartButtons = document.querySelectorAll(".atc-btn");
  if (!addToCartButtons || !addToCartButtons.length) {
    // not a product page
    throw null;
  }
  const cartWidget = document.querySelector("#cart-widget");

  for (const btn of addToCartButtons) {
    btn.addEventListener("click", function (e) {
      console.log((e.target as HTMLButtonElement).dataset.product);
    });
  }
} catch (err) {
  if (err) console.error(err);
}
