
const themeToggle = document.getElementById('themeToggle');
const themeIcon = document.getElementById('themeIcon');
const body = document.body;

// Get saved theme or default to dark
let currentTheme = localStorage.getItem('theme') || 'dark';

// Apply theme on page load
function applyTheme(theme) {
    if (theme === 'light') {
        body.setAttribute('data-theme', 'light');
        themeIcon.className = 'fas fa-sun';
        themeToggle.title = 'Switch to Dark Mode';
    } else {
        body.removeAttribute('data-theme');
        themeIcon.className = 'fas fa-moon';
        themeToggle.title = 'Switch to Light Mode';
    }
    localStorage.setItem('theme', theme);
}

// Toggle theme function
function toggleTheme() {
    currentTheme = currentTheme === 'dark' ? 'light' : 'dark';
    applyTheme(currentTheme);

    // Add rotation animation to toggle button
    themeToggle.style.transform = 'rotate(360deg)';
    setTimeout(() => {
        themeToggle.style.transform = 'rotate(0deg)';
    }, 300);
}

// Theme toggle event listener
themeToggle.addEventListener('click', toggleTheme);

// Apply saved theme on load
applyTheme(currentTheme);

const API_KEY = '99af1e52e8b504f480478eda';
const BASE_URL = 'https://v6.exchangerate-api.com/v6';
let currencies = [];
let history = JSON.parse(localStorage.getItem('conversionHistory') || '[]');

// Counter animation for statistics
function animateCounters() {
    const counters = document.querySelectorAll('[data-count]');
    counters.forEach(counter => {
        const target = parseInt(counter.dataset.count);
        const current = parseInt(counter.innerText);
        const increment = target / 100;

        if (current < target) {
            counter.innerText = Math.ceil(current + increment);
            setTimeout(() => animateCounters(), 50);
        }
    });
}

// Smooth scroll functions
function scrollToConverter() {
    document.getElementById('converter').scrollIntoView({
        behavior: 'smooth'
    });
}

function showAllCurrencies() {
    document.getElementById('currencies').scrollIntoView({
        behavior: 'smooth'
    });
    loadCurrencies();
}

// Load currencies
async function loadCurrencies() {
    const loading = document.getElementById('currencyLoading');
    const list = document.getElementById('currencyList');

    loading.style.display = 'block';
    list.innerHTML = '';

    try {
        const response = await fetch(`${BASE_URL}/${API_KEY}/codes`);
        const data = await response.json();

        if (data.result === 'success') {
            currencies = data.supported_codes;

            // Display currencies in grid
            list.innerHTML = currencies.map(([code, name]) =>
                `<div class="col-lg-3 col-md-4 col-sm-6">
                            <div class="card bg-secondary text-light mb-2">
                                <div class="card-body py-2">
                                    <strong class="text-warning">${code}</strong>
                                    <small class="d-block text-muted">${name}</small>
                                </div>
                            </div>
                        </div>`
            ).join('');

            // Load currency options for converter
            loadCurrencyOptions();
        } else {
            list.innerHTML = '<div class="col-12"><div class="alert alert-danger">Gagal memuat data mata uang.</div></div>';
        }
    } catch (error) {
        list.innerHTML = '<div class="col-12"><div class="alert alert-danger">Terjadi kesalahan saat memuat data.</div></div>';
    } finally {
        loading.style.display = 'none';
    }
}

// Load currency options for dropdowns
function loadCurrencyOptions() {
    const fromSelect = document.getElementById('fromCurrency');
    const toSelect = document.getElementById('toCurrency');

    const options = currencies.map(([code, name]) =>
        `<option value="${code}">${code} - ${name}</option>`
    ).join('');

    fromSelect.innerHTML = '<option value="">Pilih mata uang asal...</option>' + options;
    toSelect.innerHTML = '<option value="">Pilih mata uang tujuan...</option>' + options;

    // Set popular currencies as default
    fromSelect.value = 'USD';
    toSelect.value = 'IDR';
}

// Handle conversion form
document.getElementById('converterForm').addEventListener('submit', async (e) => {
    e.preventDefault();

    const from = document.getElementById('fromCurrency').value;
    const to = document.getElementById('toCurrency').value;
    const amount = parseFloat(document.getElementById('amount').value);

    if (!from || !to || !amount) {
        alert('Harap lengkapi semua field!');
        return;
    }

    const loading = document.querySelector('.loading-spinner');
    const buttonText = document.getElementById('convertButtonText');
    const resultDiv = document.getElementById('conversionResult');

    // Show loading
    loading.classList.add('show');
    buttonText.textContent = 'Sedang Konversi...';
    resultDiv.style.display = 'none';

    try {
        const response = await fetch(`${BASE_URL}/${API_KEY}/pair/${from}/${to}/${amount}`);
        const data = await response.json();

        if (data.result === 'success') {
            const result = {
                amount: amount,
                from: from,
                to: to,
                result: data.conversion_result,
                rate: data.conversion_rate,
                date: new Date().toLocaleDateString('id-ID', {
                    year: 'numeric',
                    month: 'long',
                    day: 'numeric',
                    hour: '2-digit',
                    minute: '2-digit'
                })
            };

            // Add to history
            history.unshift(result);
            if (history.length > 50) history.pop();
            localStorage.setItem('conversionHistory', JSON.stringify(history));

            // Display result
            document.getElementById('resultAmount').textContent =
                `${amount.toLocaleString('id-ID', { minimumFractionDigits: 2 })} ${from} = ${result.result.toLocaleString('id-ID', { minimumFractionDigits: 2 })} ${to}`;

            document.getElementById('resultRate').textContent =
                `Rate: 1 ${from} = ${result.rate.toFixed(4)} ${to}`;

            resultDiv.style.display = 'block';

            // Update history display
            displayHistory();

            // Success animation
            resultDiv.style.animation = 'none';
            setTimeout(() => {
                resultDiv.style.animation = 'fadeIn 0.5s ease-in';
            }, 10);

        } else {
            alert('Gagal melakukan konversi. Silakan coba lagi.');
        }
    } catch (error) {
        alert('Terjadi kesalahan saat konversi. Silakan coba lagi.');
    } finally {
        loading.classList.remove('show');
        buttonText.textContent = 'Konversi Sekarang';
    }
});

// Display conversion history
function displayHistory() {
    const historyList = document.getElementById('historyList');

    if (history.length === 0) {
        historyList.innerHTML = `
                    <div class="text-center text-muted py-4">
                        <i class="fas fa-clock fa-3x mb-3" style="color: var(--secondary-gold);"></i>
                        <p>Belum ada riwayat konversi</p>
                        <small>Lakukan konversi pertama Anda untuk melihat history di sini</small>
                    </div>
                `;
        return;
    }

    historyList.innerHTML = history.map((item, index) => `
                <div class="card bg-secondary mb-2" style="animation: fadeIn 0.3s ease-in-out ${index * 0.1}s both;">
                    <div class="card-body py-2">
                        <div class="d-flex justify-content-between align-items-center">
                            <div>
                                <strong class="text-warning">
                                    ${item.amount.toLocaleString('id-ID', { minimumFractionDigits: 2 })} ${item.from}
                                </strong>
                                <i class="fas fa-arrow-right mx-2 text-muted"></i>
                                <strong class="text-light">
                                    ${item.result.toLocaleString('id-ID', { minimumFractionDigits: 2 })} ${item.to}
                                </strong>
                            </div>
                            <small class="text-muted">${item.date}</small>
                        </div>
                        <small class="text-muted">Rate: ${item.rate.toFixed(4)}</small>
                    </div>
                </div>
            `).join('');
}

// Smooth navbar scroll effect
window.addEventListener('scroll', () => {
    const navbar = document.querySelector('.navbar');
    if (window.scrollY > 50) {
        navbar.classList.add('scrolled');
        navbar.style.backgroundColor = 'rgba(26, 26, 26, 0.95)';
        navbar.style.backdropFilter = 'blur(10px)';
    } else {
        navbar.classList.remove('scrolled');
        navbar.style.backgroundColor = '';
        navbar.style.backdropFilter = '';
    }
});

// Navbar smooth scroll
document.querySelectorAll('a[href^="#"]').forEach(anchor => {
    anchor.addEventListener('click', function (e) {
        e.preventDefault();
        const target = document.querySelector(this.getAttribute('href'));
        if (target) {
            target.scrollIntoView({
                behavior: 'smooth',
                block: 'start'
            });
        }
    });
});

// Initialize on page load
window.addEventListener('load', () => {
    // Initialize AOS with theme consideration
    AOS.init({
        duration: 1000,
        once: true
    });

    loadCurrencies();
    displayHistory();

    // Start counter animation when stats section comes into view
    const observer = new IntersectionObserver((entries) => {
        entries.forEach(entry => {
            if (entry.isIntersecting) {
                animateCounters();
                observer.unobserve(entry.target);
            }
        });
    });

    const statsSection = document.querySelector('.stats-section');
    if (statsSection) {
        observer.observe(statsSection);
    }
});

// Currency swap function
function swapCurrencies() {
    const fromSelect = document.getElementById('fromCurrency');
    const toSelect = document.getElementById('toCurrency');

    const fromValue = fromSelect.value;
    const toValue = toSelect.value;

    fromSelect.value = toValue;
    toSelect.value = fromValue;
}

// Add swap button functionality
document.addEventListener('DOMContentLoaded', () => {
    const converterForm = document.getElementById('converterForm');
    const swapButton = document.createElement('div');
    swapButton.className = 'text-center my-3';
    swapButton.innerHTML = `
                <button type="button" class="btn btn-outline-warning rounded-circle p-2" onclick="swapCurrencies()" title="Tukar Mata Uang">
                    <i class="fas fa-exchange-alt"></i>
                </button>
            `;

    const selectContainer = converterForm.querySelector('.row.g-3');
    selectContainer.parentNode.insertBefore(swapButton, selectContainer.nextSibling);
});

// Add some interactive effects
document.addEventListener('mousemove', (e) => {
    const coins = document.querySelectorAll('.floating-coin');
    coins.forEach((coin, index) => {
        const speed = 0.01 + index * 0.005;
        const x = (e.clientX * speed) * (index % 2 === 0 ? 1 : -1);
        const y = (e.clientY * speed) * (index % 2 === 0 ? 1 : -1);
        coin.style.transform = `translate(${x}px, ${y}px) rotate(${x + y}deg)`;
    });
});

// Add typing effect for hero subtitle
function typeWriter(element, text, speed = 50) {
    element.innerHTML = '';
    let i = 0;
    function type() {
        if (i < text.length) {
            element.innerHTML += text.charAt(i);
            i++;
            setTimeout(type, speed);
        }
    }
    type();
}

// Initialize typing effect
setTimeout(() => {
    const subtitle = document.querySelector('.hero-subtitle');
    const originalText = subtitle.textContent;
    typeWriter(subtitle, originalText, 30);
}, 1000);