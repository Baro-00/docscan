import './style.css';

import {
    SelectSourceDirectory,
    SelectOutputDirectory,
    Scan,
    Collect
} from '../wailsjs/go/main/App';

let documents = [];

document.querySelector('#app').innerHTML = `
<div class="container">

    <header>
        <h1>DocScan</h1>
        <p>Document discovery & triage</p>
    </header>

    <section class="group-box">
        <span class="group-title">Source</span>

        <div class="path-row">
            <input
                id="source"
                type="text"
                placeholder="C:\\Users\\user\\Documents"
            >

            <button id="sourceBrowse">
                Browse...
            </button>
        </div>

        <button id="scan" disabled>
            Scan
        </button>
    </section>

    <section class="group-box">
        <span class="group-title">Output</span>

        <div class="path-row">
            <input
                id="output"
                type="text"
                placeholder="C:\\DocScan\\Output"
            >

            <button id="outputBrowse">
                Browse...
            </button>
        </div>

        <button id="collect" disabled>
            Collect
        </button>
    </section>

    <section class="summary">
        <div>
            Documents:
            <strong id="documentCount">0</strong>
        </div>

        <div>
            Unique:
            <strong id="uniqueCount">0</strong>
        </div>

        <div>
            Duplicates:
            <strong id="duplicateCount">0</strong>
        </div>
    </section>

    <section class="results">
        <table>
            <thead>
                <tr>
                    <th>Name</th>
                    <th>Type</th>
                    <th>Size</th>
                    <th>Status</th>
                    <th>SHA-256</th>
                </tr>
            </thead>

            <tbody id="documents"></tbody>
        </table>
    </section>

    <section class="details-group">
        <span class="group-title">Document details</span>

        <div id="detailsEmpty">
            Select a document to view details.
        </div>

        <div id="details" class="details hidden">

            <div class="detail-row">
                <span>Name:</span>
                <input id="detailName" readonly>
            </div>

            <div class="detail-row">
                <span>Source:</span>
                <input id="detailSource" readonly>
            </div>

            <div class="detail-row">
                <span>Output:</span>
                <input id="detailOutput" readonly>
            </div>

            <div class="detail-grid">
                <div>
                    <span>Type:</span>
                    <strong id="detailType"></strong>
                </div>

                <div>
                    <span>Size:</span>
                    <strong id="detailSize"></strong>
                </div>

                <div>
                    <span>Status:</span>
                    <strong id="detailStatus"></strong>
                </div>
            </div>

            <div class="detail-row">
                <span>MD5:</span>
                <input id="detailMD5" class="hash-input" readonly>
            </div>

            <div class="detail-row">
                <span>SHA-256:</span>
                <input id="detailSHA256" class="hash-input" readonly>
            </div>

        </div>
    </section>

    <footer id="status">
        Ready
    </footer>

</div>
`;

const sourceInput = document.querySelector('#source');
const outputInput = document.querySelector('#output');

const sourceBrowseButton =
    document.querySelector('#sourceBrowse');

const outputBrowseButton =
    document.querySelector('#outputBrowse');

const scanButton =
    document.querySelector('#scan');

const collectButton =
    document.querySelector('#collect');

const statusElement =
    document.querySelector('#status');

sourceInput.addEventListener('input', updateButtons);
outputInput.addEventListener('input', updateButtons);

function updateButtons() {
    scanButton.disabled =
        sourceInput.value.trim() === '';

    collectButton.disabled =
        documents.length === 0 ||
        outputInput.value.trim() === '';
}

sourceBrowseButton.addEventListener('click', async () => {
    try {
        const path = await SelectSourceDirectory();

        if (!path) {
            return;
        }

        sourceInput.value = path;

        updateButtons();

        setStatus('Source directory selected');
    } catch (error) {
        setStatus(`Error: ${error}`);
    }
});

outputBrowseButton.addEventListener('click', async () => {
    try {
        const path = await SelectOutputDirectory();

        if (!path) {
            return;
        }

        outputInput.value = path;

        updateButtons();

        setStatus('Output directory selected');
    } catch (error) {
        setStatus(`Error: ${error}`);
    }
});

scanButton.addEventListener('click', async () => {
    const path = sourceInput.value.trim();

    if (!path) {
        return;
    }

    try {
        setBusy(true);
        setStatus('Scanning...');

        documents = await Scan(path);

        renderDocuments();

        setStatus(
            `Scan complete — ${documents.length} documents`
        );
    } catch (error) {
        console.error(error);

        setStatus(`Scan failed: ${error}`);
    } finally {
        setBusy(false);
        updateButtons();
    }
});

collectButton.addEventListener('click', async () => {
    const output = outputInput.value.trim();

    if (!output || documents.length === 0) {
        return;
    }

    try {
        setBusy(true);
        setStatus('Collecting documents...');

        documents = await Collect(
            documents,
            output
        );

        if (selectedDocument !== null) {
            renderDetails(documents[selectedDocument]);
        }

        renderDocuments();

        const copied = documents.filter(
            doc => doc.collectionStatus === 'COPIED'
        ).length;

        const skipped = documents.filter(
            doc => doc.collectionStatus === 'SKIPPED'
        ).length;

        setStatus(
            `Collection complete — copied: ${copied}, skipped: ${skipped}`
        );

    } catch (error) {
        console.error(error);

        setStatus(`Collection failed: ${error}`);
    } finally {
        setBusy(false);
        updateButtons();
    }
});

function setBusy(busy) {
    sourceInput.disabled = busy;
    outputInput.disabled = busy;

    sourceBrowseButton.disabled = busy;
    outputBrowseButton.disabled = busy;

    scanButton.disabled = busy;
    collectButton.disabled = busy;
}

function statusFor(doc) {
    if (doc.duplicate) {
        return 'DUPLICATE';
    }

    switch (doc.collectionStatus) {
        case 'COPIED':
            return 'COPIED';

        case 'SKIPPED':
            return 'SKIPPED';

        default:
            return 'UNIQUE';
    }
}

function renderDocuments() {
    const tbody = document.querySelector('#documents');

    tbody.innerHTML = '';

    let duplicateCount = 0;

    documents.forEach((doc, index) => {
        const row = document.createElement('tr');

        const documentStatus = statusFor(doc);

        row.dataset.index = index;

        row.innerHTML = `
            <td class="filename">
                ${escapeHtml(doc.name)}
            </td>

            <td>${escapeHtml(doc.type)}</td>

            <td>${formatSize(doc.size)}</td>

            <td>${documentStatus}</td>

            <td
                class="hash"
                title="${escapeHtml(doc.sha256)}"
            >
                ${doc.sha256.substring(0, 16)}…
            </td>
        `;

        row.addEventListener('click', () => {
            selectDocument(index);
        });

        tbody.appendChild(row);
    });

    document.querySelector('#documentCount').textContent =
        documents.length;

    document.querySelector('#duplicateCount').textContent =
        duplicateCount;

    document.querySelector('#uniqueCount').textContent =
        documents.length - duplicateCount;
}

function setStatus(message) {
    statusElement.textContent = message;
}

let selectedDocument = null;

function selectDocument(index) {
    selectedDocument = index;

    const doc = documents[index];

    document
        .querySelectorAll('#documents tr')
        .forEach(row => {
            row.classList.remove('selected');
        });

    const row = document.querySelector(
        `#documents tr[data-index="${index}"]`
    );

    if (row) {
        row.classList.add('selected');
    }

    renderDetails(doc);
}

function renderDetails(doc) {
    document.querySelector('#detailsEmpty')
        .classList.add('hidden');

    document.querySelector('#details')
        .classList.remove('hidden');

    document.querySelector('#detailName').value =
        doc.name ?? '';

    document.querySelector('#detailSource').value =
        doc.sourcePath ?? '';

    document.querySelector('#detailOutput').value =
        doc.outputPath ?? '';

    document.querySelector('#detailType').textContent =
        doc.type ?? '';

    document.querySelector('#detailSize').textContent =
        formatSize(doc.size);

    document.querySelector('#detailStatus').textContent =
        statusFor(doc);

    document.querySelector('#detailMD5').value =
        doc.md5 ?? '';

    document.querySelector('#detailSHA256').value =
        doc.sha256 ?? '';
}

function formatSize(bytes) {
    if (bytes < 1024) {
        return `${bytes} B`;
    }

    if (bytes < 1024 * 1024) {
        return `${(bytes / 1024).toFixed(1)} KB`;
    }

    if (bytes < 1024 * 1024 * 1024) {
        return `${(
            bytes / 1024 / 1024
        ).toFixed(1)} MB`;
    }

    return `${(
        bytes / 1024 / 1024 / 1024
    ).toFixed(2)} GB`;
}

function escapeHtml(value) {
    const div = document.createElement('div');

    div.textContent = value ?? '';

    return div.innerHTML;
}