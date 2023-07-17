"use strict";
var __awaiter = (this && this.__awaiter) || function (thisArg, _arguments, P, generator) {
    function adopt(value) { return value instanceof P ? value : new P(function (resolve) { resolve(value); }); }
    return new (P || (P = Promise))(function (resolve, reject) {
        function fulfilled(value) { try { step(generator.next(value)); } catch (e) { reject(e); } }
        function rejected(value) { try { step(generator["throw"](value)); } catch (e) { reject(e); } }
        function step(result) { result.done ? resolve(result.value) : adopt(result.value).then(fulfilled, rejected); }
        step((generator = generator.apply(thisArg, _arguments || [])).next());
    });
};
;
;
;
;
const buildForm = document.querySelector("#build-gate");
if (buildForm) {
    buildForm.addEventListener("submit", function (e) {
        return __awaiter(this, void 0, void 0, function* () {
            e.preventDefault();
            try {
                const input = buildForm.querySelector("#desired-width");
                const width = parseInt(input === null || input === void 0 ? void 0 : input.value);
                const res = yield fetch("/build/", {
                    method: "POST",
                    headers: {
                        "Content-Type": "application/json",
                    },
                    body: JSON.stringify({ width }),
                });
                if (res.ok) {
                    const data = (yield res.json());
                    let elems = ``;
                    for (let i = 0; i < data.length; i++) {
                        const bundle = data[i];
                        const extensions = [];
                        for (let ii = 0; ii < bundle.extensions.length; ii++) {
                            const extension = bundle.extensions[ii];
                            extensions.push({
                                id: extension.id,
                                qty: extension.qty,
                            });
                        }
                        const params = new URLSearchParams();
                        params.set("gate", JSON.stringify({ id: bundle.gates[0].id, qty: bundle.gates[0].qty }));
                        params.set("extensions", JSON.stringify(extensions));
                        const bundleURL = params.toString();
                        const productCard = (id, type, name, color, price, img) => `<div class="x-2 mb-4">
          <div class="bg-white rounded-lg overflow-hidden shadow-md">
            <a href="/${type}s/${id}"
              ><!--src="/assets/${img}"--><img src="https://via.placeholder.com/500x300"  alt="${name}" class="w-full"
            /></a>
            <div class="p-4">
              <a href="/${type}s/${id}"
                ><h3 class="font-bold mb-2">${name} ${color}</h3></a
              >
              <p class="text-gray-600 mb-4">
                This baby safety gate is perfect for keeping your baby safe in any room
                of your house.
              </p>
              <div class="flex justify-between items-center">
                <span class="text-xl font-bold">€${price}</span>
              </div>
            </div>
          </div>
        </div>`;
                        let gateHTML = ``;
                        for (const gate of bundle.gates) {
                            gateHTML += productCard(gate.id, "gate", gate.name, gate.color, gate.price, gate.img);
                        }
                        let extensionHTML = ``;
                        for (const extension of bundle.extensions) {
                            extensionHTML += productCard(extension.id, "extension", extension.name, extension.color, extension.price, extension.img);
                        }
                        const html = /*html*/ `<a id="fade-in-element" class="hidden" href="/bundles/?${bundleURL}">
          <div class="bg-white rounded-lg overflow-hidden shadow-md">
            <h3 class="font-bold mb-2">${bundle.name}</h3>
            <span class="text-xl font-bold flex">€${bundle.price}</span>
            ${gateHTML + extensionHTML}
          </div></a>`;
                        elems += html;
                    }
                    const htmlToInsert = elems;
                    console.log(elems);
                    const resultsArea = document.querySelector("#build-results");
                    if (resultsArea) {
                        resultsArea.classList.remove("hidden");
                        const container = document.querySelector("#build-results-content");
                        if (container) {
                            container.innerHTML = htmlToInsert;
                            const fadeElems = container.querySelectorAll("#fade-in-element");
                            if (fadeElems) {
                                for (let x = 0; x < (fadeElems === null || fadeElems === void 0 ? void 0 : fadeElems.length); x++) {
                                    const fElem = fadeElems[x];
                                    yield new Promise(function (resolve) {
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
            }
            catch (err) {
                console.log(err);
            }
        });
    });
}
else {
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
            console.log(e.target.dataset.product);
        });
    }
}
catch (err) {
    if (err)
        console.error(err);
}
