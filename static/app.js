// ============================================
// Dynamic DB Manager - Frontend Logic
// ============================================

const API = '/api';
let currentDB = '';        // currently selected database
let editContext = null;    // { dbName, tableName, pkCol, pkVal, columns }

// ---------- Navigation ----------

function navigate(page) {
    document.querySelectorAll('.page').forEach(p => p.classList.remove('active'));
    document.getElementById('page-' + page)?.classList.add('active');

    document.querySelectorAll('.nav-btn').forEach(b => b.classList.remove('active'));
    event?.target.closest('.nav-btn')?.classList.add('active');

    // Pre-load data for pages that need it
    if (['add-column', 'insert', 'browse'].includes(page) && currentDB) {
        loadTableDropdowns();
    }
}

// ---------- Toast notifications ----------

function toast(msg, type = 'success') {
    const el = document.getElementById('toast');
    el.textContent = msg;
    el.className = 'toast ' + type;
    clearTimeout(el._timer);
    el._timer = setTimeout(() => el.classList.add('hidden'), 4000);
}

// ---------- API helpers ----------

async function api(path, opts = {}) {
    try {
        const res = await fetch(API + path, {
            headers: { 'Content-Type': 'application/json' },
            ...opts,
        });
        const json = await res.json();
        if (!json.success) throw new Error(json.error || 'Unknown error');
        return json;
    } catch (err) {
        toast(err.message, 'error');
        throw err;
    }
}

// ---------- Dashboard: list databases ----------

async function loadDatabases() {
    try {
        const res = await api('/databases');
        const dbs = res.data || [];
        const container = document.getElementById('db-list');
        if (dbs.length === 0) {
            container.innerHTML = '<p class="muted">No databases yet. Create one!</p>';
            return;
        }
        container.innerHTML = dbs.map(name => `
            <div class="db-card ${name === currentDB ? 'selected' : ''}" onclick="selectDB('${name}')">
                <span class="db-icon">🗄️</span>
                <span class="db-name">${name}</span>
            </div>
        `).join('');
    } catch (_) { /* toast already shown */ }
}

function selectDB(name) {
    currentDB = name;
    document.getElementById('current-db-badge').textContent = '📦 ' + name;
    document.getElementById('current-db-badge').className = 'active-db';
    loadDatabases(); // refresh highlights
    toast(`Switched to database: ${name}`);
}

// ---------- Create Database ----------

async function createDatabase() {
    const input = document.getElementById('new-db-name');
    const name = input.value.trim();
    if (!name) return toast('Enter a database name', 'error');
    await api('/databases', { method: 'POST', body: JSON.stringify({ name }) });
    input.value = '';
    toast(`Database '${name}' created`);
    loadDatabases();
}

// ---------- Load table dropdowns ----------

async function loadTableDropdowns() {
    if (!currentDB) return;
    try {
        const res = await api(`/databases/${currentDB}/tables`);
        const tables = res.data?.tables || [];
        const opts = '<option value="">Select table...</option>' +
            tables.map(t => `<option value="${t}">${t}</option>`).join('');
        ['addcol-table', 'insert-table', 'browse-table'].forEach(id => {
            const el = document.getElementById(id);
            if (el) el.innerHTML = opts;
        });
    } catch (_) {}
}

// ---------- Create Table ----------

function addColumnRow() {
    const builder = document.getElementById('column-builder');
    const row = document.createElement('div');
    row.className = 'col-row';
    row.innerHTML = `
        <input type="text" placeholder="Column name" class="col-name">
        <select class="col-type">
            <option value="">Type...</option>
            <option value="INT AUTO_INCREMENT PRIMARY KEY">ID (Auto PK)</option>
            <option value="VARCHAR(255)">VARCHAR(255)</option>
            <option value="VARCHAR(100)">VARCHAR(100)</option>
            <option value="INT">INT</option>
            <option value="BIGINT">BIGINT</option>
            <option value="FLOAT">FLOAT</option>
            <option value="DOUBLE">DOUBLE</option>
            <option value="DECIMAL(10,2)">DECIMAL(10,2)</option>
            <option value="TEXT">TEXT</option>
            <option value="DATE">DATE</option>
            <option value="DATETIME">DATETIME</option>
            <option value="BOOLEAN">BOOLEAN</option>
        </select>
        <button onclick="this.parentElement.remove()" class="btn btn-icon btn-danger" title="Remove">✕</button>
    `;
    builder.appendChild(row);
}

async function createTable() {
    if (!currentDB) return toast('Select a database first', 'error');
    const tableName = document.getElementById('new-table-name').value.trim();
    if (!tableName) return toast('Enter a table name', 'error');

    const columns = {};
    let valid = false;
    document.querySelectorAll('#column-builder .col-row').forEach(row => {
        const name = row.querySelector('.col-name').value.trim();
        const type = row.querySelector('.col-type').value;
        if (name && type) { columns[name] = type; valid = true; }
    });
    if (!valid) return toast('Add at least one column', 'error');

    await api(`/databases/${currentDB}/tables`, {
        method: 'POST',
        body: JSON.stringify({ name: tableName, columns }),
    });
    document.getElementById('new-table-name').value = '';
    toast(`Table '${tableName}' created`);
}

// ---------- Add Column ----------

async function addColumn() {
    if (!currentDB) return toast('Select a database first', 'error');
    const table = document.getElementById('addcol-table').value;
    const name = document.getElementById('addcol-name').value.trim();
    const type = document.getElementById('addcol-type').value;
    if (!table || !name || !type) return toast('Fill in all fields', 'error');

    await api(`/databases/${currentDB}/tables/${table}/columns`, {
        method: 'POST',
        body: JSON.stringify({ name, type }),
    });
    document.getElementById('addcol-name').value = '';
    toast(`Column '${name}' added to '${table}'`);
}

// ---------- Insert Record ----------

async function loadInsertForm() {
    const table = document.getElementById('insert-table').value;
    const container = document.getElementById('insert-form-fields');
    const submitBtn = document.getElementById('insert-submit');
    if (!table || !currentDB) { container.innerHTML = ''; submitBtn.style.display = 'none'; return; }

    try {
        const res = await api(`/databases/${currentDB}/tables/${table}/schema`);
        const cols = res.data?.columns || [];
        container.innerHTML = cols.map(col => {
            const inputType = getInputType(col.type);
            return `
                <div class="insert-field">
                    <label>${col.name} <span class="muted">(${col.type})</span></label>
                    <input type="${inputType}" data-col="${col.name}" placeholder="Enter value...">
                </div>
            `;
        }).join('');
        submitBtn.style.display = 'block';
    } catch (_) {}
}

async function submitInsert() {
    const table = document.getElementById('insert-table').value;
    if (!table || !currentDB) return;

    const data = {};
    let hasData = false;
    document.querySelectorAll('#insert-form-fields input').forEach(inp => {
        const val = inp.value.trim();
        if (val) { data[inp.dataset.col] = val; hasData = true; }
    });
    if (!hasData) return toast('Enter at least one value', 'error');

    await api(`/databases/${currentDB}/tables/${table}/records`, {
        method: 'POST',
        body: JSON.stringify(data),
    });
    document.querySelectorAll('#insert-form-fields input').forEach(i => i.value = '');
    toast('Record inserted');
}

// ---------- Browse / View Records ----------

async function loadRecords() {
    const table = document.getElementById('browse-table').value;
    const area = document.getElementById('records-area');
    if (!table || !currentDB) return toast('Select a database and table', 'error');

    try {
        const res = await api(`/databases/${currentDB}/tables/${table}/records`);
        const cols = res.data?.columns || [];
        const records = res.data?.records || [];

        if (records.length === 0) {
            area.innerHTML = '<div class="no-data">📭 No records found</div>';
            return;
        }

        // Detect primary key (first column whose type contains 'int' and 'auto_increment' or simply the first column)
        const pkCol = cols.length > 0 ? cols[0].name : null;

        let html = `<div class="record-count">${records.length} record(s) found</div>`;
        html += '<div class="table-wrap"><table class="data-table"><thead><tr>';
        cols.forEach(c => html += `<th>${c.name}</th>`);
        html += '<th>Actions</th></tr></thead><tbody>';

        records.forEach(rec => {
            html += '<tr>';
            cols.forEach(c => {
                const val = rec[c.name];
                html += `<td>${val !== null && val !== undefined ? val : '<span class="muted">NULL</span>'}</td>`;
            });
            const pkVal = rec[pkCol];
            const recJSON = encodeURIComponent(JSON.stringify(rec));
            const colsJSON = encodeURIComponent(JSON.stringify(cols));
            html += `<td class="actions">
                <button class="btn btn-sm btn-ghost" onclick="openEditModal('${table}','${pkCol}','${pkVal}', '${recJSON}', '${colsJSON}')">✏️</button>
                <button class="btn btn-sm btn-danger" onclick="deleteRecord('${table}','${pkCol}','${pkVal}')">🗑️</button>
            </td>`;
            html += '</tr>';
        });
        html += '</tbody></table></div>';
        area.innerHTML = html;
    } catch (_) {}
}

// ---------- Edit modal ----------

function openEditModal(table, pkCol, pkVal, recJSON, colsJSON) {
    const record = JSON.parse(decodeURIComponent(recJSON));
    const cols = JSON.parse(decodeURIComponent(colsJSON));
    editContext = { table, pkCol, pkVal, cols };

    const fields = document.getElementById('edit-fields');
    fields.innerHTML = cols.map(col => {
        const val = record[col.name] !== null && record[col.name] !== undefined ? record[col.name] : '';
        const disabled = col.name === pkCol ? 'disabled' : '';
        return `
            <div class="insert-field">
                <label>${col.name} <span class="muted">(${col.type})</span></label>
                <input type="${getInputType(col.type)}" data-col="${col.name}" value="${val}" ${disabled}>
            </div>
        `;
    }).join('');

    document.getElementById('edit-modal').classList.remove('hidden');
}

function closeEditModal() {
    document.getElementById('edit-modal').classList.add('hidden');
    editContext = null;
}

async function submitEdit() {
    if (!editContext) return;
    const data = {};
    let changed = false;
    document.querySelectorAll('#edit-fields input:not([disabled])').forEach(inp => {
        const val = inp.value.trim();
        if (val !== '') { data[inp.dataset.col] = val; changed = true; }
    });
    if (!changed) return toast('No changes made', 'error');

    await api(`/databases/${currentDB}/tables/${editContext.table}/records`, {
        method: 'PUT',
        body: JSON.stringify({
            data,
            condition_col: editContext.pkCol,
            condition_val: editContext.pkVal,
        }),
    });
    closeEditModal();
    toast('Record updated');
    loadRecords();
}

// ---------- Delete ----------

async function deleteRecord(table, pkCol, pkVal) {
    if (!confirm('Delete this record?')) return;
    await api(`/databases/${currentDB}/tables/${table}/records`, {
        method: 'DELETE',
        body: JSON.stringify({ condition_col: pkCol, condition_val: pkVal }),
    });
    toast('Record deleted');
    loadRecords();
}

// ---------- Sample DB ----------

async function createSampleDB() {
    await api('/sample', { method: 'POST' });
    toast('RealEstate sample database created');
    loadDatabases();
}

// ---------- Helpers ----------

function getInputType(mysqlType) {
    const t = mysqlType.toLowerCase();
    if (t.includes('int') || t.includes('float') || t.includes('double') || t.includes('decimal')) return 'number';
    if (t.includes('datetime') || t.includes('timestamp')) return 'datetime-local';
    if (t.includes('date')) return 'date';
    return 'text';
}

// ---------- Init ----------

document.addEventListener('DOMContentLoaded', () => {
    loadDatabases();
});