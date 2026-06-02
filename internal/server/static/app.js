/*
 * Copyright 2026 The Planwright Authors
 * SPDX-License-Identifier: Apache-2.0
 */

const planInput = document.querySelector("#plan-input");
const form = document.querySelector("#plan-form");
const loadExample = document.querySelector("#load-example");
const statusBox = document.querySelector("#status");
const diagnosticsList = document.querySelector("#diagnostics");
const nodesTable = document.querySelector("#nodes");
const edgesTable = document.querySelector("#edges");
const preview = document.querySelector("#preview");
const tabButtons = Array.from(document.querySelectorAll("[data-preview]"));

const state = {
  selectedPreview: "security",
  reports: {},
  terraformFiles: [],
  mermaidFiles: [],
};

form.addEventListener("submit", async (event) => {
  event.preventDefault();
  await validatePlan();
});

loadExample.addEventListener("click", async () => {
  setStatus("Loading example.", "");
  try {
    const response = await fetch("/api/example", { headers: { Accept: "application/json" } });
    const payload = await response.json();
    if (!response.ok) {
      throw new Error(payload.error || "Example could not be loaded.");
    }
    planInput.value = payload.plan;
    await validatePlan();
  } catch (error) {
    setStatus(error.message, "error");
  }
});

for (const button of tabButtons) {
  button.addEventListener("click", () => {
    state.selectedPreview = button.dataset.preview;
    for (const tab of tabButtons) {
      tab.setAttribute("aria-selected", String(tab === button));
    }
    renderPreview();
  });
}

async function validatePlan() {
  setStatus("Validating plan.", "");
  try {
    const response = await fetch("/api/validate", {
      method: "POST",
      headers: {
        Accept: "application/json",
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ plan: planInput.value }),
    });
    const payload = await response.json();
    if (!response.ok && payload.error) {
      clearResults();
      setStatus(payload.error, "error");
      return;
    }
    state.reports = payload.reports || {};
    state.terraformFiles = payload.terraform_files || [];
    state.mermaidFiles = payload.mermaid_files || [];
    renderDiagnostics(payload.diagnostics || []);
    renderNodes((payload.graph && payload.graph.nodes) || []);
    renderEdges((payload.graph && payload.graph.edges) || []);
    renderPreview();
    if (payload.blocking) {
      setStatus("Validation found blocking diagnostics.", "error");
    } else {
      setStatus("Validation passed. Previews are ready.", "ok");
    }
  } catch (error) {
    clearResults();
    setStatus(error.message, "error");
  }
}

function renderDiagnostics(diagnostics) {
  diagnosticsList.replaceChildren();
  if (diagnostics.length === 0) {
    const item = document.createElement("li");
    item.textContent = "No diagnostics.";
    diagnosticsList.append(item);
    return;
  }
  for (const diagnostic of diagnostics) {
    const item = document.createElement("li");
    item.className = diagnostic.severity || "";
    item.textContent = `${upper(diagnostic.severity)} ${diagnostic.code}: ${diagnostic.message || ""}${diagnostic.fix ? " Fix: " + diagnostic.fix : ""}`;
    diagnosticsList.append(item);
  }
}

function renderNodes(nodes) {
  nodesTable.replaceChildren();
  for (const node of nodes) {
    appendRow(nodesTable, [node.id || "", node.kind || ""]);
  }
}

function renderEdges(edges) {
  edgesTable.replaceChildren();
  for (const edge of edges) {
    const flow = [edge.kind, edge.protocol && edge.port ? `${edge.protocol}/${edge.port}` : ""].filter(Boolean).join(" ");
    appendRow(edgesTable, [edge.from || "", edge.to || "", flow]);
  }
}

function appendRow(tbody, values) {
  const row = document.createElement("tr");
  for (const value of values) {
    const cell = document.createElement("td");
    cell.textContent = value;
    row.append(cell);
  }
  tbody.append(row);
}

function renderPreview() {
  if (state.selectedPreview === "terraform") {
    preview.textContent = joinFiles(state.terraformFiles);
    return;
  }
  if (state.selectedPreview === "mermaid") {
    preview.textContent = joinFiles(state.mermaidFiles);
    return;
  }
  preview.textContent = state.reports[state.selectedPreview] || "No preview available.";
}

function joinFiles(files) {
  if (!files || files.length === 0) {
    return "No files generated.";
  }
  return files.map((file) => `# ${file.path}\n\n${file.content}`).join("\n\n");
}

function clearResults() {
  state.reports = {};
  state.terraformFiles = [];
  state.mermaidFiles = [];
  diagnosticsList.replaceChildren();
  nodesTable.replaceChildren();
  edgesTable.replaceChildren();
  preview.textContent = "";
}

function setStatus(message, className) {
  statusBox.className = `status ${className}`.trim();
  statusBox.textContent = message;
}

function upper(value) {
  return String(value || "info").toUpperCase();
}

loadExample.click();
