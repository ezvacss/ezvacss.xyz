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
    const isLocalDev = window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1';

    // Device Detection & Mobile Input Placeholders
    const isMobileDevice = window.innerWidth <= 768 || ('ontouchstart' in window);
    if (isMobileDevice) {
        if (urlInput) {
            urlInput.placeholder = "Tap or paste long URL...";
        }
        const codeInput = document.getElementById('code-input');
        if (codeInput) {
            codeInput.placeholder = "Tap or paste short URL or code...";
        }
        const reportCodeInput = document.getElementById('report-code-input');
        if (reportCodeInput) {
            reportCodeInput.placeholder = "Tap or paste short link or code...";
        }
    }

    // Global keyboard shortcut: Press '/' to focus input
    window.addEventListener('keydown', (e) => {
        const activeElem = document.activeElement;
        const isTyping = activeElem && (activeElem.tagName === 'INPUT' || activeElem.tagName === 'TEXTAREA' || activeElem.isContentEditable);

        if (!isTyping && e.key === '/') {
            e.preventDefault();
            const input = document.getElementById('url-input') || document.getElementById('code-input');
            if (input) {
                input.focus();
                if (typeof input.select === 'function') input.select();
                input.scrollIntoView({ behavior: 'smooth', block: 'center' });
            }
        }
    });

    const captureLink = document.querySelector('.nav-link[data-section="capture"]');
    if (captureLink) {
        captureLink.addEventListener('click', (e) => {
            e.preventDefault();
            window.scrollTo({ top: 0, behavior: 'smooth' });
            history.pushState(null, '', '#capture');
        });
    }

    const tryFreeBtn = document.getElementById('try-free-cta-btn');
    if (tryFreeBtn) {
        tryFreeBtn.addEventListener('click', (e) => {
            e.preventDefault();
            const input = document.getElementById('url-input');
            if (input) {
                input.scrollIntoView({ behavior: 'smooth', block: 'center' });
                input.focus();
            }
        });
    }

    const normalizePath = (p) => p.replace(/\/+$/, '') || '/';
    const currentPath = normalizePath(window.location.pathname);

    const footerLinks = document.querySelectorAll('.footer-links a');
    footerLinks.forEach((link) => {
        try {
            const linkPath = normalizePath(new URL(link.href, window.location.origin).pathname);
            if (linkPath === currentPath) {
                link.classList.add('active-glow');
            }
        } catch (e) {}
    });

    const passwordField = document.getElementById('password') || document.getElementById('password-input');
    const passwordConfirmField = document.getElementById('password-confirm');
    const passwordMatchMsg = document.getElementById('password-match-msg');

    if (passwordField && passwordConfirmField && passwordMatchMsg) {
        function checkPasswordMatch() {
            if (!passwordConfirmField.value) {
                passwordMatchMsg.style.display = 'none';
                return;
            }
            if (passwordField.value === passwordConfirmField.value) {
                passwordMatchMsg.textContent = 'Passwords match';
                passwordMatchMsg.style.color = 'var(--color-success)';
            } else {
                passwordMatchMsg.textContent = "Passwords don't match";
                passwordMatchMsg.style.color = 'var(--color-danger)';
            }
            passwordMatchMsg.style.display = 'block';
        }

        passwordField.addEventListener('input', checkPasswordMatch);
        passwordConfirmField.addEventListener('input', checkPasswordMatch);
    }

    function showError(msg) {
        if (!errorMessage) return;

        if (errorMessageText) {
            errorMessageText.textContent = msg;
        } else {
            errorMessage.textContent = msg;
        }
        errorMessage.style.display = 'block';
        errorMessage.classList.add('visible');
        if (resultsBox) {
            resultsBox.style.display = 'none';
        }
    }

    function hideError() {
        if (!errorMessage) return;

        errorMessage.style.display = 'none';
        errorMessage.classList.remove('visible');
    }


    if (shortenForm) {
        shortenForm.addEventListener('submit', async (e) => {
            e.preventDefault();

            let buttonTextTimer;

            function setButtonText(element, text) {
                if (!element || element.textContent === text) return;

                clearTimeout(buttonTextTimer);
                element.classList.add('is-changing');

                buttonTextTimer = setTimeout(() => {
                    element.textContent = text;
                    element.classList.remove('is-changing');
                }, 150);
            }

            const originalUrl = urlInput.value.trim();
            if (!originalUrl) return;


            let testStr = originalUrl;
            if (!/^https?:\/\//i.test(testStr)) {
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

            const captchaContainer = document.getElementById('captcha-container');
            const isCaptchaVisible = captchaContainer && captchaContainer.style.display !== 'none';
            const turnstileResponse = document.querySelector('[name="cf-turnstile-response"]');
            const turnstileToken = turnstileResponse ? turnstileResponse.value : '';

            if (isCaptchaVisible && !turnstileToken && !isLocalDev) {
                showError('Please complete the captcha verification.');
                return;
            }

            if (btnText) btnText.textContent = 'Shortening...';
            shortenBtn.disabled = true;
            shortenBtn.setAttribute('aria-busy', 'true');

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

                const qrApiUrl = `https://api.qrserver.com/v1/create-qr-code/?size=180x180&data=${encodeURIComponent(data.short_url)}`;
                qrCodeImage.src = qrApiUrl;

                hideError();
                resultsBox.style.display = 'block';
                resultsBox.scrollIntoView({behavior: 'smooth', block: 'nearest'});

                urlInput.value = '';

            } catch (error) {
                console.error('Error:', error);
                showError(error.message || 'Something went wrong. Please check your internet connection or try again.');
                if (typeof turnstile !== 'undefined') {
                    turnstile.reset();
                }
            } finally {
                if (btnText) btnText.textContent = 'Shorten';
                shortenBtn.disabled = false;
                shortenBtn.setAttribute('aria-busy', 'false');
                if (resultsBox && resultsBox.style.display === 'block' && typeof turnstile !== 'undefined') {
                    turnstile.reset();
                }
            }
        });
    }

    if (urlInput) {
        urlInput.addEventListener('input', hideError);
    }

    if (copyUrlBtn) {
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
    }

    // QR Code Actions: Download (HD 600x600), Share, Print (HD)
    const downloadQrBtn = document.getElementById('download-qr-btn');
    if (downloadQrBtn && qrCodeImage) {
        downloadQrBtn.addEventListener('click', async () => {
            const targetUrl = shortenedUrlText ? (shortenedUrlText.href || shortenedUrlText.textContent) : '';
            if (!targetUrl) return;

            // Fetch HD 600x600 PNG image for high-quality downloads & prints
            const hdQrUrl = `https://api.qrserver.com/v1/create-qr-code/?size=600x600&data=${encodeURIComponent(targetUrl)}`;
            try {
                const response = await fetch(hdQrUrl);
                const blob = await response.blob();
                const blobUrl = URL.createObjectURL(blob);
                const a = document.createElement('a');
                a.href = blobUrl;
                a.download = 'ezvacss-qr-code-hd.png';
                document.body.appendChild(a);
                a.click();
                document.body.removeChild(a);
                URL.revokeObjectURL(blobUrl);
            } catch (err) {
                console.error('Failed to download HD QR code image:', err);
                window.open(hdQrUrl, '_blank');
            }
        });
    }

    const shareQrBtn = document.getElementById('share-qr-btn');
    if (shareQrBtn) {
        shareQrBtn.addEventListener('click', async () => {
            const shortUrl = shortenedUrlText ? (shortenedUrlText.href || shortenedUrlText.textContent) : window.location.href;
            if (navigator.share) {
                try {
                    await navigator.share({
                        title: 'ezvacss.xyz Short Link',
                        text: 'Check out this shortened link created with ezvacss.xyz:',
                        url: shortUrl
                    });
                } catch (err) {
                    if (err.name !== 'AbortError') {
                        console.error('Share failed:', err);
                    }
                }
            } else {
                try {
                    await navigator.clipboard.writeText(shortUrl);
                    const span = shareQrBtn.querySelector('span');
                    if (span) {
                        const origText = span.textContent;
                        span.textContent = 'Copied!';
                        setTimeout(() => { span.textContent = origText; }, 2000);
                    }
                } catch (e) {
                    console.error('Copy fallback failed:', e);
                }
            }
        });
    }

    const printQrBtn = document.getElementById('print-qr-btn');
    if (printQrBtn) {
        printQrBtn.addEventListener('click', () => {
            const shortUrl = shortenedUrlText ? (shortenedUrlText.href || shortenedUrlText.textContent) : '';
            if (!shortUrl) return;

            const hdQrUrl = `https://api.qrserver.com/v1/create-qr-code/?size=600x600&data=${encodeURIComponent(shortUrl)}`;
            const printWin = window.open('', '_blank', 'width=600,height=700');
            if (printWin) {
                printWin.document.write(`
                    <!DOCTYPE html>
                    <html>
                    <head>
                        <title>Print QR Code | ezvacss.xyz</title>
                        <style>
                            body { font-family: system-ui, -apple-system, sans-serif; text-align: center; padding: 3rem; background: #fff; color: #111; }
                            .card { border: 2px solid #e2e8f0; border-radius: 16px; padding: 2.5rem; display: inline-block; max-width: 400px; box-shadow: 0 10px 25px rgba(0,0,0,0.1); }
                            h2 { font-size: 1.6rem; margin: 0 0 0.5rem 0; color: #0f172a; letter-spacing: -0.02em; }
                            p { color: #64748b; font-size: 0.95rem; margin-bottom: 1.75rem; }
                            img { width: 220px; height: 220px; border-radius: 8px; border: 1px solid #cbd5e1; padding: 0.5rem; background: #fff; }
                            .url { font-family: monospace; font-size: 1.15rem; font-weight: 700; color: #4f46e5; margin-top: 1.5rem; word-break: break-all; }
                            .footer { margin-top: 1.5rem; font-size: 0.8rem; color: #94a3b8; }
                        </style>
                    </head>
                    <body>
                        <div class="card">
                            <h2>ezvacss.xyz</h2>
                            <p>Scan with any phone camera to access link</p>
                            <img src="${hdQrUrl}" alt="High-Res QR Code">
                            <div class="url">${shortUrl}</div>
                            <div class="footer">Created with ezvacss.xyz link shortener</div>
                        </div>
                        <script>
                            window.onload = () => { window.print(); window.close(); };
                        <\/script>
                    </body>
                    </html>
                `);
                printWin.document.close();
            }
        });
    }
    
    async function checkAuth() {
        const userDisplay = document.getElementById('user-display');
        const dashboardLink = document.getElementById('dashboard-link');
        const logoutBtn = document.getElementById('logout-btn');
        const authBtn = document.getElementById('auth-btn');
        const registerBtn = document.getElementById('register-nav-btn');

        function loadTurnstileScript() {
            if (!document.getElementById('cf-turnstile-script')) {
                const script = document.createElement('script');
                script.id = 'cf-turnstile-script';
                script.src = 'https://challenges.cloudflare.com/turnstile/v0/api.js';
                script.async = true;
                script.defer = true;
                document.head.appendChild(script);
            }
        }

        try {
            const res = await fetch('/api/user');
            const captchaContainer = document.getElementById('captcha-container');
            if (res.ok) {
                const user = await res.json();
                if (user.authenticated) {
                    if (userDisplay) {
                        userDisplay.textContent = `@${user.username}`;
                        userDisplay.style.display = 'inline-block';
                    }
                    if (dashboardLink) dashboardLink.style.display = 'inline-block';
                    if (logoutBtn) logoutBtn.style.display = 'inline-block';
                    if (authBtn) authBtn.style.display = 'none';
                    if (registerBtn) registerBtn.style.display = 'none';
                    if (captchaContainer) captchaContainer.style.display = 'none';
                } else {
                    if (userDisplay) userDisplay.style.display = 'none';
                    if (dashboardLink) dashboardLink.style.display = 'none';
                    if (logoutBtn) logoutBtn.style.display = 'none';
                    if (authBtn) authBtn.style.display = 'inline-flex';
                    if (registerBtn) registerBtn.style.display = 'inline-flex';
                    if (captchaContainer) {
                        if (isLocalDev) {
                            captchaContainer.style.display = 'none';
                        } else {
                            captchaContainer.style.display = 'flex';
                            loadTurnstileScript();
                        }
                    }
                }
            } else {
                if (userDisplay) userDisplay.style.display = 'none';
                if (dashboardLink) dashboardLink.style.display = 'none';
                if (logoutBtn) logoutBtn.style.display = 'none';
                if (authBtn) authBtn.style.display = 'inline-flex';
                if (registerBtn) registerBtn.style.display = 'inline-flex';
                if (captchaContainer) {
                    if (isLocalDev) {
                        captchaContainer.style.display = 'none';
                    } else {
                        captchaContainer.style.display = 'flex';
                        loadTurnstileScript();
                    }
                }
            }
        } catch (err) {
            console.error('Auth check failed:', err);
            if (userDisplay) userDisplay.style.display = 'none';
            if (dashboardLink) dashboardLink.style.display = 'none';
            if (logoutBtn) logoutBtn.style.display = 'none';
            if (authBtn) authBtn.style.display = 'inline-flex';
            if (registerBtn) registerBtn.style.display = 'inline-flex';
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
