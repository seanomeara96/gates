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
                    const data = yield res.json();
                    const elems = [];
                    for (let i = 0; i < data.length; i++) {
                        const bundle = data[i];
                        const html = /*html*/ `
          <a id="fade-in-element" class="w-full md:w-1/2 lg:w-1/3 px-2 mb-4 hidden" href="/gates/{{.Id}}">
            <div class="bg-white rounded-lg overflow-hidden shadow-md">
                <img src="https://via.placeholder.com/500x300" alt="Baby Safety Gate" class="w-full">
                <div class="p-4">
                    <h3 class="font-bold mb-2">${bundle.gate.name}${bundle.gate.color}${bundle.extensions.length ? " &amp; " + bundle.extensions.reduce((a, c) => a + c.qty, 0) + " Extensions" : ""}</h3>
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
