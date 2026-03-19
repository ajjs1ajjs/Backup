// Patch for users.html - Fix modal size and add edit functionality

// 1. Fix modal CSS (line ~196)
const modalCSS = `
            .modal {
                background: var(--bg-card);
                border-radius: 1rem;
                padding: 2rem;
                width: 100%;
                max-width: 800px;
                max-height: 90vh;
                overflow-y: auto;
                border: 1px solid var(--border);
            }
`;

// 2. Fix editUser function (line ~739)
function editUser(id) {
    const user = users.find(u => u.id === id);
    if (!user) return;

    // Fill form
    document.getElementById('user-username').value = user.username;
    document.getElementById('user-password').value = '';
    document.getElementById('user-password').placeholder = 'Залиште пустим щоб не змінювати';
    document.getElementById('user-fullname').value = user.full_name || '';
    document.getElementById('user-email').value = user.email || '';
    document.getElementById('user-role').value = user.role || 'readonly';
    document.getElementById('user-enabled').checked = user.enabled;

    // Store editing ID
    window.editingUserId = id;

    // Change button text
    const btn = document.querySelector('#add-modal .btn-primary');
    if (btn) btn.textContent = '💾 Зберегти зміни';

    openAddModal();
}

console.log('Users.html patch applied!');
