document.addEventListener('DOMContentLoaded', () => {
    const shortenForm = document.getElementById('shorten-form');
    const urlInput = document.getElementById('url-input');
    const shortenBtn = document.getElementById('shorten-btn');
    const resultsBox = document.getElementById('results-box-container');
    const shortenedUrlText = document.getElementById('shortened-url-text');
    const originalUrlText = document.getElementById('original-url-text');
    const copyUrlBtn = document.getElementById('copy-url-btn');
    const copyBtnIcon = document.getElementById('copy-btn-icon');
    const qrCodeImage = document.getElementById('qr-code-image');

    
    const errorMessage = document.getElementById('error-message');
    const errorMessageText = document.getElementById('error-message-text');

    function showError(msg) {
        if (errorMessage && errorMessageText) {
            errorMessageText.textContent = msg;
            errorMessage.style.display = 'block';
        }
        if (resultsBox) {
            resultsBox.style.display = 'none';
        }
    }

    function hideError() {
        if (errorMessage) {
            errorMessage.style.display = 'none';
        }
    }

    
    shortenForm.addEventListener('submit', async (e) => {
        e.preventDefault();
        hideError();
        
        const originalUrl = urlInput.value.trim();
        if (!originalUrl) return;

        
        let testStr = originalUrl;
        if (!/^https?:\/\
            testStr = 'http://' + testStr;
        }
        let parsedUrl;
        try {
            parsedUrl = new URL(testStr);
        } catch (_) {
            showError('Please enter a valid URL (e.g. google.com).');
            return;
        }

        const host = parsedUrl.hostname.toLowerCase();
        if (!host.includes('.') && host !== 'localhost') {
            showError('Please enter a valid web address (e.g. google.com).');
            return;
        }

        if (host === window.location.hostname || host === 'ezvacss.xyz') {
            showError('Cannot shorten links targeting our own domain.');
            return;
        }

        
        const btnText = shortenBtn.querySelector('span');
        const btnIcon = shortenBtn.querySelector('i');
        const originalBtnText = btnText.textContent;

        const captchaContainer = document.getElementById('captcha-container');
        const isCaptchaVisible = captchaContainer && captchaContainer.style.display !== 'none';
        const turnstileResponse = document.querySelector('[name="cf-turnstile-response"]');
        const turnstileToken = turnstileResponse ? turnstileResponse.value : '';

        if (isCaptchaVisible && !turnstileToken) {
            showError('Please complete the captcha verification.');
            return;
        }

        btnText.textContent = 'Shortening...';
        shortenBtn.disabled = true;

        try {
            const response = await fetch('/shorten', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ 
                    url: originalUrl,
                    turnstile_token: turnstileToken
                }),
            });

            if (!response.ok) {
                const text = await response.text();
                throw new Error(text || 'Failed to shorten URL');
            }

            const data = await response.json();
            
            
            originalUrlText.textContent = originalUrl;
            shortenedUrlText.textContent = data.short_url;
            shortenedUrlText.href = data.short_url;

            
            
            const qrApiUrl = `https://api.qrserver.com/v1/create-qr-code/?size=150x150&data=${encodeURIComponent(data.short_url)}`;
            qrCodeImage.src = qrApiUrl;

            
            resultsBox.style.display = 'block';
            resultsBox.scrollIntoView({ behavior: 'smooth', block: 'nearest' });

            
            urlInput.value = '';

        } catch (error) {
            console.error('Error:', error);
            showError(error.message || 'Something went wrong. Please check your internet connection or try again.');
            if (typeof turnstile !== 'undefined') {
                turnstile.reset();
            }
        } finally {
            btnText.textContent = originalBtnText;
            shortenBtn.disabled = false;
            if (resultsBox && resultsBox.style.display === 'block' && typeof turnstile !== 'undefined') {
                turnstile.reset();
            }
        }
    });

    
    copyUrlBtn.addEventListener('click', async () => {
        const urlToCopy = shortenedUrlText.textContent;
        try {
            await navigator.clipboard.writeText(urlToCopy);
            
            
            copyBtnIcon.setAttribute('data-lucide', 'check');
            copyUrlBtn.style.borderColor = 'var(--color-success)';
            copyUrlBtn.style.color = 'var(--color-success)';
            lucide.createIcons();

            setTimeout(() => {
                copyBtnIcon.setAttribute('data-lucide', 'copy');
                copyUrlBtn.style.borderColor = '';
                copyUrlBtn.style.color = '';
                lucide.createIcons();
            }, 2000);

        } catch (err) {
            console.error('Failed to copy text: ', err);
        }
    });

    
    async function checkAuth() {
        const userDisplay = document.getElementById('user-display');
        const dashboardLink = document.getElementById('dashboard-link');
        const logoutBtn = document.getElementById('logout-btn');
        const authBtn = document.getElementById('auth-btn');
        const registerBtn = document.getElementById('register-nav-btn');

        try {
            const res = await fetch('/api/user');
            const captchaContainer = document.getElementById('captcha-container');
            if (res.ok) {
                const user = await res.json();
                userDisplay.textContent = `@${user.username}`;
                userDisplay.style.display = 'inline-block';
                dashboardLink.style.display = 'inline-block';
                logoutBtn.style.display = 'inline-block';
                authBtn.style.display = 'none';
                if (registerBtn) registerBtn.style.display = 'none';
                if (captchaContainer) captchaContainer.style.display = 'none';
            } else {
                userDisplay.style.display = 'none';
                dashboardLink.style.display = 'none';
                logoutBtn.style.display = 'none';
                authBtn.style.display = 'inline-flex';
                if (registerBtn) registerBtn.style.display = 'inline-flex';
                if (captchaContainer) captchaContainer.style.display = 'flex';
            }
        } catch (err) {
            console.error('Auth check failed:', err);
        } finally {
            const navMenuBar = document.getElementById('nav-menu-bar');
            if (navMenuBar) {
                navMenuBar.style.opacity = '1';
            }
        }
    }

    
    const logoutBtn = document.getElementById('logout-btn');
    if (logoutBtn) {
        logoutBtn.addEventListener('click', async () => {
            try {
                const res = await fetch('/api/logout', { method: 'POST' });
                if (res.ok) {
                    window.location.reload();
                }
            } catch (err) {
                console.error('Logout failed:', err);
            }
        });
    }

    
    checkAuth();
});
