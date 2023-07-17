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
        const resultsArea = document.querySelector("#build-results");
        if (resultsArea) {
          resultsArea.classList.remove("hidden");

          const container = document.querySelector("#build-results-content");
          if (container) {
            container.innerHTML = await res.text();

            const fadeElems = container.querySelectorAll("#fade-in-element");
            if (fadeElems) {
              for (let x = 0; x < fadeElems?.length; x++) {
                const fElem = fadeElems[x];
                await (function fadeAnimation() {
                  return new Promise(function (resolve) {
                    fElem.classList.remove("hidden");
                    fElem.addEventListener("animationend", function () {
                      fElem.classList.remove("fade-in");
                      resolve(null);
                    });
                    fElem.classList.add("fade-in");
                  });
                })();
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
