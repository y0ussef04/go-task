// Global variables
let currentSection = 'database';
let currentDatabase = '';

// API base URL
const API_BASE = '/api';

// Utility functions
function showStatus(elementId, message, type = 'info') {
    const element = document.getElementById(elementId);
    element.textContent = message;
    element.className = `status ${type}`;
    element.style.display = 'block';

    // Auto-hide after 5 seconds
    setTimeout(() => {
        element.style.display = 'none';
    }, 5000);
}

function clearStatus(elementId) {
    document.getElementById(elementId).style.display = 'none';
}

function showSection(sectionName) {
    // Hide all sections
    document.querySelectorAll('.section').forEach(section => {
        section.classList.add('hidden');
    });

    // Show selected section
    document.getElementById(`${sectionName}-section`).classList.remove('hidden');
    currentSection = sectionName;

    // Clear any previous status messages
    clearStatus('db-status');
    clearStatus('table-status');
    clearStatus('data-status');
    clearStatus('sample-status');
}

// Database operations
async function createDatabase() {
    const dbName = document.getElementById('db-name').value.trim();
    if (!dbName) {
        showStatus('db-status', 'Please enter a database name', 'error');
        return;
    }

    try {
        const response = await fetch(`${API_BASE}/databases`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ name: dbName }),
        });

        const result = await response.json();

        if (response.ok) {
            showStatus('db-status', result.message, 'success');
            document.getElementById('db-name').value = '';
        } else {
            showStatus('db-status', result.error, 'error');
        }
    } catch (error) {
        showStatus('db-status', 'Network error: ' + error.message, 'error');
    }
}

async function useDatabase() {
    const dbName = document.getElementById('use-db-dropdown').value.trim();
    if (!dbName) {
        showStatus('db-status', 'Please select a database', 'error');
        return;
    }

    try {
        const response = await fetch(`${API_BASE}/databases/use`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ name: dbName }),
        });

        const result = await response.json();

        if (response.ok) {
            showStatus('db-status', result.message, 'success');
            currentDatabase = dbName; // Store current database
            // Reload tables in all dropdowns
            loadTablesToInsert();
            loadTablesToView();
            loadTablesToUpdate();
            loadTablesToDelete();
        } else {
            showStatus('db-status', result.error, 'error');
        }
    } catch (error) {
        showStatus('db-status', 'Network error: ' + error.message, 'error');
    }
}

// Load databases for the Use Database section
async function loadDatabasesToUse() {
    try {
        const response = await fetch(`${API_BASE}/databases/list`);
        const result = await response.json();

        if (response.ok && result.data) {
            const dropdown = document.getElementById('use-db-dropdown');
            dropdown.innerHTML = '<option value="">Select a database</option>';
            
            result.data.forEach(db => {
                const option = document.createElement('option');
                option.value = db;
                option.textContent = db;
                dropdown.appendChild(option);
            });
            
            showStatus('db-status', 'Databases loaded', 'success');
        } else {
            showStatus('db-status', 'Failed to load databases', 'error');
        }
    } catch (error) {
        showStatus('db-status', 'Network error: ' + error.message, 'error');
    }
}

// Load tables for quick insert
async function loadTablesForQuickInsert() {
    if (!currentDatabase) {
        const dropdown = document.getElementById('quick-insert-table');
        dropdown.innerHTML = '<option value="">Select a database first</option>';
        return;
    }

    try {
        const response = await fetch(`${API_BASE}/databases/${currentDatabase}`);
        const result = await response.json();

        if (response.ok && result.data && result.data.tables) {
            const dropdown = document.getElementById('quick-insert-table');
            dropdown.innerHTML = '<option value="">Choose a table to insert into</option>';
            
            result.data.tables.forEach(table => {
                const option = document.createElement('option');
                option.value = table;
                option.textContent = table;
                dropdown.appendChild(option);
            });
        }
    } catch (error) {
        console.error('Error loading tables:', error);
    }
}

// Generate quick insert form when table is selected
async function generateQuickInsertForm() {
    const tableName = document.getElementById('quick-insert-table').value.trim();
    if (!tableName || !currentDatabase) {
        document.getElementById('quick-insert-form').style.display = 'none';
        return;
    }

    try {
        const response = await fetch(`${API_BASE}/databases/${currentDatabase}/tables/${tableName}/schema`);
        const result = await response.json();

        if (response.ok && result.data && result.data.columns) {
            const container = document.getElementById('quick-insert-fields');
            container.innerHTML = '';

            result.data.columns.forEach(col => {
                const fieldDiv = document.createElement('div');
                fieldDiv.className = 'field-input';
                
                // Create appropriate input type based on column type
                let inputType = 'text';
                let placeholder = `Enter ${col.type.toLowerCase()} value`;
                
                if (col.type.toLowerCase().includes('int') || col.type.toLowerCase().includes('float') || col.type.toLowerCase().includes('double') || col.type.toLowerCase().includes('decimal')) {
                    inputType = 'number';
                    placeholder = `Enter number (${col.type})`;
                } else if (col.type.toLowerCase().includes('date')) {
                    inputType = 'date';
                    placeholder = 'Select date';
                } else if (col.type.toLowerCase().includes('datetime') || col.type.toLowerCase().includes('timestamp')) {
                    inputType = 'datetime-local';
                    placeholder = 'Select date and time';
                }
                
                fieldDiv.innerHTML = `
                    <label>${col.name} (${col.type}):</label>
                    <input type="${inputType}" name="${col.name}" placeholder="${placeholder}" data-column="${col.name}">
                `;
                container.appendChild(fieldDiv);
            });

            document.getElementById('quick-insert-form').style.display = 'block';
            showStatus('data-status', `Ready to insert into '${tableName}'`, 'success');
        } else {
            showStatus('data-status', 'Failed to load table schema', 'error');
        }
    } catch (error) {
        showStatus('data-status', 'Network error: ' + error.message, 'error');
    }
}

// Submit quick insert form
async function submitQuickInsert() {
    const tableName = document.getElementById('quick-insert-table').value.trim();
    if (!tableName) {
        showStatus('data-status', 'Please select a table', 'error');
        return;
    }

    if (!currentDatabase) {
        showStatus('data-status', 'Please select a database first', 'error');
        return;
    }

    const data = {};
    const inputs = document.querySelectorAll('#quick-insert-fields input');
    let hasData = false;
    
    inputs.forEach(input => {
        const column = input.getAttribute('data-column');
        const value = input.value.trim();
        if (value !== '') {
            data[column] = value;
            hasData = true;
        }
    });

    if (!hasData) {
        showStatus('data-status', 'Please enter at least one value', 'error');
        return;
    }

    try {
        const response = await fetch(`${API_BASE}/databases/${currentDatabase}/tables/${tableName}/records`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(data),
        });

        const result = await response.json();

        if (response.ok) {
            showStatus('data-status', result.message, 'success');
            // Clear form
            inputs.forEach(input => input.value = '');
        } else {
            showStatus('data-status', result.error, 'error');
        }
    } catch (error) {
        showStatus('data-status', 'Network error: ' + error.message, 'error');
    }
}

// Load schema automatically when table is selected for insert
async function loadTableSchemaForInsert() {
    const tableName = document.getElementById('insert-table-dropdown').value.trim();
    if (!tableName) {
        document.getElementById('insert-fields').style.display = 'none';
        document.getElementById('insert-btn').style.display = 'none';
        return;
    }

    try {
        const response = await fetch(`${API_BASE}/tables/${tableName}/schema`);
        const result = await response.json();

        if (response.ok && result.data && result.data.columns) {
            const container = document.getElementById('insert-fields');
            container.innerHTML = '<h4>Enter Data:</h4>';

            result.data.columns.forEach(col => {
                const fieldDiv = document.createElement('div');
                fieldDiv.className = 'field-input';
                fieldDiv.innerHTML = `
                    <label>${col.name} (${col.type}):</label>
                    <input type="text" data-column="${col.name}" placeholder="Enter ${col.type.toLowerCase()} value">
                `;
                container.appendChild(fieldDiv);
            });

            container.style.display = 'block';
            document.getElementById('insert-btn').style.display = 'block';
            showStatus('data-status', `Ready to insert into '${tableName}'`, 'success');
        } else {
            showStatus('data-status', 'Failed to load table schema', 'error');
        }
    } catch (error) {
        showStatus('data-status', 'Network error: ' + error.message, 'error');
    }
}

// Load tables for View operations
async function loadTablesToView() {
    if (!currentDatabase) {
        const dropdown = document.getElementById('view-table-dropdown');
        dropdown.innerHTML = '<option value="">Select a database first</option>';
        return;
    }

    try {
        const response = await fetch(`${API_BASE}/databases/${currentDatabase}`);
        const result = await response.json();

        if (response.ok && result.data && result.data.tables) {
            const dropdown = document.getElementById('view-table-dropdown');
            dropdown.innerHTML = '<option value="">Select a table</option>';
            
            result.data.tables.forEach(table => {
                const option = document.createElement('option');
                option.value = table;
                option.textContent = table;
                dropdown.appendChild(option);
            });
        }
    } catch (error) {
        console.error('Error loading tables:', error);
    }
}

// Load tables for Update operations
async function loadTablesToUpdate() {
    try {
        const response = await fetch(`${API_BASE}/tables/list`);
        const result = await response.json();

        if (response.ok && result.data) {
            const dropdown = document.getElementById('update-table-dropdown');
            dropdown.innerHTML = '<option value="">Select a table</option>';
            
            result.data.forEach(table => {
                const option = document.createElement('option');
                option.value = table;
                option.textContent = table;
                dropdown.appendChild(option);
            });
        }
    } catch (error) {
        console.error('Error loading tables:', error);
    }
}

// Load schema automatically when table is selected for update
async function loadTableSchemaForUpdate() {
    const tableName = document.getElementById('update-table-dropdown').value.trim();
    if (!tableName) {
        document.getElementById('update-fields').style.display = 'none';
        document.getElementById('update-condition').style.display = 'none';
        document.getElementById('update-btn').style.display = 'none';
        return;
    }

    try {
        const response = await fetch(`${API_BASE}/tables/${tableName}/schema`);
        const result = await response.json();

        if (response.ok && result.data && result.data.columns) {
            const container = document.getElementById('update-fields');
            container.innerHTML = '<h4>Update Data:</h4>';

            result.data.columns.forEach(col => {
                const fieldDiv = document.createElement('div');
                fieldDiv.className = 'field-input';
                fieldDiv.innerHTML = `
                    <label>${col.name} (${col.type}):</label>
                    <input type="text" data-column="${col.name}" placeholder="New ${col.type.toLowerCase()} value (leave empty to keep current)">
                `;
                container.appendChild(fieldDiv);
            });

            container.style.display = 'block';
            document.getElementById('update-condition').style.display = 'block';
            document.getElementById('update-btn').style.display = 'block';
            showStatus('data-status', `Ready to update '${tableName}'`, 'success');
        } else {
            showStatus('data-status', 'Failed to load table schema', 'error');
        }
    } catch (error) {
        showStatus('data-status', 'Network error: ' + error.message, 'error');
    }
}

// Load tables for Delete operations
async function loadTablesToDelete() {
    try {
        const response = await fetch(`${API_BASE}/tables/list`);
        const result = await response.json();

        if (response.ok && result.data) {
            const dropdown = document.getElementById('delete-table-dropdown');
            dropdown.innerHTML = '<option value="">Select a table</option>';
            
            result.data.forEach(table => {
                const option = document.createElement('option');
                option.value = table;
                option.textContent = table;
                dropdown.appendChild(option);
            });
        }
    } catch (error) {
        console.error('Error loading tables:', error);
    }
}

// Table operations
function addColumnInput() {
    const container = document.getElementById('columns-container');
    const columnInput = document.createElement('div');
    columnInput.className = 'column-input';
    columnInput.innerHTML = `
        <input type="text" placeholder="Column name" class="col-name" required>
        <select class="col-type" required>
            <option value="">Select data type</option>
            <option value="VARCHAR(255)">VARCHAR(255)</option>
            <option value="VARCHAR(50)">VARCHAR(50)</option>
            <option value="VARCHAR(100)">VARCHAR(100)</option>
            <option value="INT">INT</option>
            <option value="BIGINT">BIGINT</option>
            <option value="FLOAT">FLOAT</option>
            <option value="DOUBLE">DOUBLE</option>
            <option value="DECIMAL(10,2)">DECIMAL(10,2)</option>
            <option value="DATE">DATE</option>
            <option value="DATETIME">DATETIME</option>
            <option value="TIMESTAMP">TIMESTAMP</option>
            <option value="TEXT">TEXT</option>
            <option value="BOOLEAN">BOOLEAN</option>
            <option value="TINYINT(1)">TINYINT(1)</option>
            <option value="BLOB">BLOB</option>
        </select>
        <button onclick="removeColumn(this)" class="btn-danger">Remove</button>
    `;
    container.appendChild(columnInput);
}

function removeColumn(button) {
    button.parentElement.remove();
}

async function createTable() {
    const tableName = document.getElementById('table-name').value.trim();
    if (!tableName) {
        showStatus('table-status', 'Please enter a table name', 'error');
        return;
    }

    const columns = {};
    const columnInputs = document.querySelectorAll('.column-input');
    let hasValidColumns = false;

    columnInputs.forEach(input => {
        const colName = input.querySelector('.col-name').value.trim();
        const colType = input.querySelector('.col-type').value.trim();

        if (colName && colType) {
            columns[colName] = colType;
            hasValidColumns = true;
        }
    });

    if (!hasValidColumns) {
        showStatus('table-status', 'Please add at least one valid column', 'error');
        return;
    }

    try {
        const response = await fetch(`${API_BASE}/tables`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                name: tableName,
                columns: columns,
            }),
        });

        const result = await response.json();

        if (response.ok) {
            showStatus('table-status', result.message, 'success');
            document.getElementById('table-name').value = '';
            document.getElementById('columns-container').innerHTML = `
                <div class="column-input">
                    <input type="text" placeholder="Column name" class="col-name" required>
                    <select class="col-type" required>
                        <option value="">Select data type</option>
                        <option value="VARCHAR(255)">VARCHAR(255)</option>
                        <option value="VARCHAR(50)">VARCHAR(50)</option>
                        <option value="VARCHAR(100)">VARCHAR(100)</option>
                        <option value="INT">INT</option>
                        <option value="BIGINT">BIGINT</option>
                        <option value="FLOAT">FLOAT</option>
                        <option value="DOUBLE">DOUBLE</option>
                        <option value="DECIMAL(10,2)">DECIMAL(10,2)</option>
                        <option value="DATE">DATE</option>
                        <option value="DATETIME">DATETIME</option>
                        <option value="TIMESTAMP">TIMESTAMP</option>
                        <option value="TEXT">TEXT</option>
                        <option value="BOOLEAN">BOOLEAN</option>
                        <option value="TINYINT(1)">TINYINT(1)</option>
                        <option value="BLOB">BLOB</option>
                    </select>
                    <button onclick="removeColumn(this)" class="btn-danger">Remove</button>
                </div>
            `;
        } else {
            showStatus('table-status', result.error, 'error');
        }
    } catch (error) {
        showStatus('table-status', 'Network error: ' + error.message, 'error');
    }
}

async function addColumn() {
    const tableName = document.getElementById('add-col-table').value.trim();
    const colName = document.getElementById('add-col-name').value.trim();
    const colType = document.getElementById('add-col-type').value.trim();

    if (!tableName || !colName || !colType) {
        showStatus('table-status', 'Please fill in all fields', 'error');
        return;
    }

    try {
        const response = await fetch(`${API_BASE}/tables/${tableName}/columns`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                name: colName,
                type: colType,
            }),
        });

        const result = await response.json();

        if (response.ok) {
            showStatus('table-status', result.message, 'success');
            document.getElementById('add-col-table').value = '';
            document.getElementById('add-col-name').value = '';
            document.getElementById('add-col-type').value = '';
        } else {
            showStatus('table-status', result.error, 'error');
        }
    } catch (error) {
        showStatus('table-status', 'Network error: ' + error.message, 'error');
    }
}

// Data operations
async function loadTableSchema(operation) {
    const dropdownId = `${operation}-table-dropdown`;
    const tableName = document.getElementById(dropdownId).value.trim();
    if (!tableName) {
        showStatus('data-status', 'Please select a table', 'error');
        return;
    }

    try {
        const response = await fetch(`${API_BASE}/tables/${tableName}/schema`);
        const result = await response.json();

        if (response.ok) {
            const container = document.getElementById(`${operation}-fields`);
            container.innerHTML = '';

            result.columns.forEach(col => {
                const fieldDiv = document.createElement('div');
                fieldDiv.className = 'field-input';
                fieldDiv.innerHTML = `
                    <label>${col.name} (${col.type}):</label>
                    <input type="text" data-column="${col.name}" placeholder="Enter ${col.type.toLowerCase()} value">
                `;
                container.appendChild(fieldDiv);
            });

            showStatus('data-status', `Schema loaded for table '${tableName}'`, 'success');
        } else {
            showStatus('data-status', result.error, 'error');
        }
    } catch (error) {
        showStatus('data-status', 'Network error: ' + error.message, 'error');
    }
}

async function insertRecord() {
    const tableName = document.getElementById('insert-table-dropdown').value.trim();
    if (!tableName) {
        showStatus('data-status', 'Please select a table', 'error');
        return;
    }

    const data = {};
    const inputs = document.querySelectorAll('#insert-fields input');
    inputs.forEach(input => {
        const column = input.getAttribute('data-column');
        const value = input.value.trim();
        if (value !== '') {
            data[column] = value;
        }
    });

    if (Object.keys(data).length === 0) {
        showStatus('data-status', 'Please enter at least one value', 'error');
        return;
    }

    try {
        const response = await fetch(`${API_BASE}/tables/${tableName}/records`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(data),
        });

        const result = await response.json();

        if (response.ok) {
            showStatus('data-status', result.message, 'success');
            // Clear inputs
            inputs.forEach(input => input.value = '');
        } else {
            showStatus('data-status', result.error, 'error');
        }
    } catch (error) {
        showStatus('data-status', 'Network error: ' + error.message, 'error');
    }
}

async function viewRecords() {
    const tableName = document.getElementById('view-table-dropdown').value.trim();
    if (!tableName) {
        showStatus('data-status', 'Please select a table', 'error');
        return;
    }

    if (!currentDatabase) {
        showStatus('data-status', 'Please select a database first', 'error');
        return;
    }

    try {
        const response = await fetch(`${API_BASE}/databases/${currentDatabase}/tables/${tableName}/records`);
        const result = await response.json();

        if (response.ok && result.data && result.data.records) {
            const container = document.getElementById('records-display');
            const records = result.data.records;

            if (records.length > 0) {
                // Create a nice table
                let html = `<h4>📊 Records in '${tableName}' (${records.length} records)</h4>`;
                html += '<div class="table-container"><table class="data-table">';
                
                // Header
                const columns = Object.keys(records[0]);
                html += '<thead><tr>';
                columns.forEach(col => {
                    html += `<th>${col}</th>`;
                });
                html += '</tr></thead>';
                
                // Body
                html += '<tbody>';
                records.forEach(record => {
                    html += '<tr>';
                    columns.forEach(col => {
                        const value = record[col] || 'NULL';
                        html += `<td>${value}</td>`;
                    });
                    html += '</tr>';
                });
                html += '</tbody></table></div>';
                
                container.innerHTML = html;
                showStatus('data-status', `Found ${records.length} records`, 'success');
            } else {
                container.innerHTML = '<div class="no-data">📭 No records found in this table</div>';
                showStatus('data-status', 'No records found', 'info');
            }
        } else {
            showStatus('data-status', result.error || 'Failed to load records', 'error');
            document.getElementById('records-display').innerHTML = '';
        }
    } catch (error) {
        showStatus('data-status', 'Network error: ' + error.message, 'error');
    }
}

async function updateRecord() {
    const tableName = document.getElementById('update-table-dropdown').value.trim();
    const condition = document.getElementById('update-condition').value.trim();

    if (!tableName || !condition) {
        showStatus('data-status', 'Please select table and enter condition', 'error');
        return;
    }

    const data = {};
    const inputs = document.querySelectorAll('#update-fields input');
    inputs.forEach(input => {
        const column = input.getAttribute('data-column');
        const value = input.value.trim();
        if (value !== '') {
            data[column] = value;
        }
    });

    if (Object.keys(data).length === 0) {
        showStatus('data-status', 'Please enter at least one value to update', 'error');
        return;
    }

    try {
        const response = await fetch(`${API_BASE}/tables/${tableName}/records`, {
            method: 'PUT',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                data: data,
                condition: condition,
            }),
        });

        const result = await response.json();

        if (response.ok) {
            showStatus('data-status', result.message, 'success');
            // Clear inputs
            inputs.forEach(input => input.value = '');
            document.getElementById('update-condition').value = '';
        } else {
            showStatus('data-status', result.error, 'error');
        }
    } catch (error) {
        showStatus('data-status', 'Network error: ' + error.message, 'error');
    }
}

async function deleteRecord() {
    const tableName = document.getElementById('delete-table-dropdown').value.trim();
    const condition = document.getElementById('delete-condition').value.trim();

    if (!tableName || !condition) {
        showStatus('data-status', 'Please select table and enter condition', 'error');
        return;
    }

    if (!confirm('Are you sure you want to delete records? This action cannot be undone.')) {
        return;
    }

    try {
        const response = await fetch(`${API_BASE}/tables/${tableName}/records`, {
            method: 'DELETE',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ condition: condition }),
        });

        const result = await response.json();

        if (response.ok) {
            showStatus('data-status', result.message, 'success');
            document.getElementById('delete-condition').value = '';
        } else {
            showStatus('data-status', result.error, 'error');
        }
    } catch (error) {
        showStatus('data-status', 'Network error: ' + error.message, 'error');
    }
}

async function createSampleDB() {
    try {
        const response = await fetch(`${API_BASE}/sample`, {
            method: 'POST',
        });

        const result = await response.json();

        if (response.ok) {
            showStatus('sample-status', result.message, 'success');
        } else {
            showStatus('sample-status', result.error, 'error');
        }
    } catch (error) {
        showStatus('sample-status', 'Network error: ' + error.message, 'error');
    }
}

// Initialize the application
document.addEventListener('DOMContentLoaded', function() {
    showSection('database');
    // Load databases and tables on page load
    loadDatabasesToUse();
    loadTablesForQuickInsert();
    loadTablesToView();
    loadTablesToUpdate();
    loadTablesToDelete();
});