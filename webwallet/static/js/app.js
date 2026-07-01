// app.js - Blockchain Web Wallet JavaScript

// Utility Functions
function formatAddress(address) {
    if (!address) return '';
    if (address.length <= 16) return address;
    return address.substring(0, 8) + '...' + address.substring(address.length - 8);
}

function formatDate(timestamp) {
    if (!timestamp) return 'N/A';
    const date = new Date(timestamp * 1000);
    return date.toLocaleString();
}

function copyToClipboard(text) {
    if (navigator.clipboard && navigator.clipboard.writeText) {
        navigator.clipboard.writeText(text).then(() => {
            showNotification('Copied to clipboard!', 'success');
        }).catch(() => {
            // Fallback
            fallbackCopy(text);
        });
    } else {
        fallbackCopy(text);
    }
}

function fallbackCopy(text) {
    const textarea = document.createElement('textarea');
    textarea.value = text;
    textarea.style.position = 'fixed';
    textarea.style.opacity = '0';
    document.body.appendChild(textarea);
    textarea.select();
    try {
        document.execCommand('copy');
        showNotification('Copied to clipboard!', 'success');
    } catch (err) {
        showNotification('Failed to copy', 'error');
    }
    document.body.removeChild(textarea);
}

function showNotification(message, type = 'info') {
    // Check if notification container exists
    let container = document.getElementById('notification-container');
    if (!container) {
        container = document.createElement('div');
        container.id = 'notification-container';
        container.style.cssText = `
            position: fixed;
            top: 20px;
            right: 20px;
            z-index: 9999;
            max-width: 400px;
        `;
        document.body.appendChild(container);
    }

    const notification = document.createElement('div');
    const colors = {
        success: '#4CAF50',
        error: '#f44336',
        info: '#2196F3',
        warning: '#ff9800'
    };
    
    notification.style.cssText = `
        background: ${colors[type] || colors.info};
        color: white;
        padding: 15px 20px;
        margin-bottom: 10px;
        border-radius: 8px;
        box-shadow: 0 4px 6px rgba(0,0,0,0.1);
        animation: slideIn 0.3s ease;
        font-family: 'Inter', sans-serif;
        font-size: 14px;
    `;
    notification.textContent = message;
    
    container.appendChild(notification);
    
    // Auto-remove after 3 seconds
    setTimeout(() => {
        notification.style.animation = 'slideOut 0.3s ease';
        setTimeout(() => {
            notification.remove();
        }, 300);
    }, 3000);
}

// Add animation styles
const styleSheet = document.createElement('style');
styleSheet.textContent = `
    @keyframes slideIn {
        from {
            transform: translateX(100%);
            opacity: 0;
        }
        to {
            transform: translateX(0);
            opacity: 1;
        }
    }
    @keyframes slideOut {
        from {
            transform: translateX(0);
            opacity: 1;
        }
        to {
            transform: translateX(100%);
            opacity: 0;
        }
    }
`;
document.head.appendChild(styleSheet);

// Wallet Functions
async function createWallet() {
    try {
        const response = await fetch('/api/createwallet');
        const data = await response.json();
        
        if (data.success) {
            showNotification('Wallet created successfully!', 'success');
            // Redirect to wallet page
            window.location.href = `/wallet?address=${data.data.address}`;
        } else {
            showNotification('Failed to create wallet: ' + data.message, 'error');
        }
    } catch (error) {
        showNotification('Error creating wallet: ' + error.message, 'error');
    }
}

async function getBalance(address) {
    if (!address) {
        showNotification('Please enter a wallet address', 'warning');
        return;
    }
    
    try {
        const response = await fetch(`/api/balance?address=${address}`);
        const data = await response.json();
        
        if (data.success) {
            showNotification(`Balance: ${data.data.balance} coins`, 'info');
            return data.data.balance;
        } else {
            showNotification('Error: ' + data.message, 'error');
            return null;
        }
    } catch (error) {
        showNotification('Error getting balance: ' + error.message, 'error');
        return null;
    }
}

async function sendCoins(from, to, amount) {
    if (!from || !to || !amount || amount <= 0) {
        showNotification('Please fill in all fields correctly', 'warning');
        return;
    }
    
    try {
        const response = await fetch('/api/send', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                from: from,
                to: to,
                amount: parseInt(amount)
            })
        });
        
        const data = await response.json();
        
        if (data.success) {
            showNotification(`Successfully sent ${amount} coins!`, 'success');
            showNotification(`Transaction ID: ${data.data.txid}`, 'info');
            // Refresh the page after a delay
            setTimeout(() => {
                window.location.reload();
            }, 2000);
        } else {
            showNotification('Failed to send coins: ' + data.message, 'error');
        }
    } catch (error) {
        showNotification('Error sending coins: ' + error.message, 'error');
    }
}

async function mineBlock(address) {
    if (!address) {
        showNotification('No wallet address found', 'warning');
        return;
    }
    
    try {
        const response = await fetch(`/api/mine?address=${address}`);
        const data = await response.json();
        
        if (data.success) {
            showNotification(`Block mined! Received ${data.data.reward} coins!`, 'success');
            // Refresh after a delay
            setTimeout(() => {
                window.location.reload();
            }, 2000);
        } else {
            showNotification('Mining failed: ' + data.message, 'error');
        }
    } catch (error) {
        showNotification('Error mining: ' + error.message, 'error');
    }
}

async function getBlockchainData() {
    try {
        const response = await fetch('/api/blocks');
        const data = await response.json();
        
        if (data.success) {
            return data.data;
        } else {
            showNotification('Failed to get blockchain data', 'error');
            return null;
        }
    } catch (error) {
        showNotification('Error getting blockchain data: ' + error.message, 'error');
        return null;
    }
}

// Search functionality
function searchAddress() {
    const input = document.getElementById('searchInput');
    const address = input.value.trim();
    
    if (!address) {
        showNotification('Please enter an address', 'warning');
        return;
    }
    
    window.location.href = `/wallet?address=${address}`;
}

// Auto-refresh balance
function autoRefresh(interval = 30000) {
    setInterval(() => {
        const addressElement = document.getElementById('walletAddress');
        if (addressElement) {
            const address = addressElement.textContent.trim();
            if (address) {
                getBalance(address);
            }
        }
    }, interval);
}

// Initialize on page load
document.addEventListener('DOMContentLoaded', function() {
    // Add search functionality if search input exists
    const searchInput = document.getElementById('searchInput');
    if (searchInput) {
        searchInput.addEventListener('keypress', function(e) {
            if (e.key === 'Enter') {
                searchAddress();
            }
        });
    }
    
    // Add copy functionality to address elements
    document.querySelectorAll('.address-value, .address-text').forEach(el => {
        el.addEventListener('click', function() {
            copyToClipboard(this.textContent.trim());
        });
        el.style.cursor = 'pointer';
    });
    
    // Auto-refresh if on wallet page
    if (window.location.pathname.includes('/wallet')) {
        autoRefresh(30000);
    }
});

// Export functions for use in HTML
window.createWallet = createWallet;
window.getBalance = getBalance;
window.sendCoins = sendCoins;
window.mineBlock = mineBlock;
window.copyToClipboard = copyToClipboard;
window.searchAddress = searchAddress;
window.formatAddress = formatAddress;