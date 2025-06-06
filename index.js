/**
 * Copyright (C) 2024  Eric Cornelissen
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

const input = document.getElementById("workflow-input");
const conservative = document.getElementById("option-conservative");

async function loadWasm() {
  const go = new Go();

  let result;
  const request = fetch("ades.wasm");
  if (typeof WebAssembly.instantiateStreaming === "function") {
    result = await WebAssembly.instantiateStreaming(request, go.importObject);
  } else {
    const response = await request;
    const mod = await response.arrayBuffer();
    result = await WebAssembly.instantiate(mod, go.importObject);
  }

  go.run(result.instance);
}

const htmlEncode = {
  ruleId: (ruleId) => {
    const link = `https://github.com/ericcornelissen/ades/blob/main/RULES.md#${ruleId}`;

    const a = document.createElement("a");
    a.setAttribute("href", link);
    a.setAttribute("rel", "noopener");
    a.setAttribute("target", "_blank");
    a.innerText = ruleId;
    return a;
  },
  violation: (violation) => {
    const li = document.createElement("li");

    li.appendChild(document.createTextNode("["));
    li.appendChild(htmlEncode.ruleId(violation.ruleId));
    li.appendChild(document.createTextNode("]"));

    if (violation.job) {
      const job = document.createElement("code");
      job.innerText = violation.job;
      li.appendChild(document.createTextNode(" In job '"));
      li.appendChild(job);
      li.appendChild(document.createTextNode("',"));
    }

    const step = document.createElement("code");
    step.innerText = violation.step;
    li.appendChild(document.createTextNode(" step '"));
    li.appendChild(step);
    li.appendChild(document.createTextNode("',"));

    const found = document.createElement("code");
    found.innerText = violation.problem;
    li.appendChild(document.createTextNode(" found '"));
    li.appendChild(found);
    li.appendChild(document.createTextNode("'."));

    return li;
  }
}

function runAnalysis() {
  const working = document.createElement("p");
  working.classList.add("working");
  working.innerText = "Working...";

  setResult(working);

  const source = getSource();
  const options = getOptions();
  if (globalThis.ades) ades(source, options);
}

function getSource() {
  return input.value.trim();
}

function getOptions() {
  return {
    conservative: conservative.checked,
  };
}

function showError(summary, full) {
  const title = document.createElement("span");
  title.classList.add("title");
  title.innerText = "Error:";

  const what = document.createElement("span");
  what.classList.add("what");
  what.innerText = summary;

  const details = document.createElement("p");
  details.classList.add("details");
  details.innerText = full;

  const error = document.createElement("div");
  error.classList.add("error");
  error.appendChild(title);
  error.appendChild(document.createTextNode(" "));
  error.appendChild(what);
  error.appendChild(details);

  setResult(error);
}

function showResult(violations) {
  const count = violations.length;
  if (count === 0) {
    const text = document.createElement("span");
    text.innerText = "No problems detected";

    setResult(text);
  } else {
    const title = document.createElement("span");
    title.classList.add("title");
    title.innerText = `Found ${count} ${count === 1 ? "problem" : "problems"}`;

    const ul = document.createElement("ul");
    for (const li of violations.map(htmlEncode.violation)) {
      ul.appendChild(li);
    }

    setResult(title, ul);
  }
}

function setResult(...children) {
  const results = document.getElementById("results");
  const parent = results.parentNode;
  parent.removeChild(results);

  const newResults = document.createElement("div");
  newResults.setAttribute("id", "results");
  for (const child of children) {
    newResults.appendChild(child);
  }

  parent.appendChild(newResults);
}

function main() {
  window.getSource = getSource;
  window.showError = showError;
  window.showResult = showResult;

  input.addEventListener("input", runAnalysis);
  conservative.addEventListener("click", runAnalysis);

  loadWasm();
}

main();
