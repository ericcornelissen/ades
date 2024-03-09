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
const results = document.getElementById("results");

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
    return `<a href="${link}" rel="noopener" target="_blank">${ruleId}</a>`;
  },
  violation: (violation) => {
    const ruleId = `[${htmlEncode.ruleId(violation.ruleId)}]`;
    const job = violation.job ? `In job '<code>${violation.job}</code>',` : "";
    const step = `step '<code>${violation.step}</code>',`;
    const problem = `found '<code>${violation.problem}</code>'`;
    return `<li>${ruleId} ${job} ${step} ${problem}.</li>`
  }
}

function runAnalysis() {
  results.innerHTML = `<p class="working">Working...</p>`;

  const source = getSource();
  ades(source);
}

function getSource() {
  return input.value.trim();
}

function showError(what, details) {
  results.innerHTML = `<div class="error"><span class="title">Error:</span> <span class="what">${what}</span><p class="details">${details}`;
}

function showResult(violations) {
  const count = violations.length;
  if (count === 0) {
    results.innerHTML = "No problems detected";
  } else {
    const listItems = violations.map(htmlEncode.violation).join("");
    const problems = count === 1 ? "problem" : "problems";
    results.innerHTML = `<div class="result"><span class="title">Found ${count} ${problems}</span><ul>${listItems}</ul></div>`;
  }
}

function main() {
  window.getSource = getSource;
  window.showError = showError;
  window.showResult = showResult;

  input.addEventListener("keyup", runAnalysis);

  loadWasm();
}

main();
