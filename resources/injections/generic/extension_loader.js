// Convenience aliases
const CUTmainWindow = PLACEHOLDER_MAINWINDOW;
const CUTwebView = PLACEHOLDER_WEBVIEW;

const CUTfs = require('fs');
const CUTpath = require('path');
const CUTelectron = require("electron");
const CUTsession = CUTelectron.session;

// Clear session cache to prevent stale SPA from loading before extensions
CUTsession.defaultSession.clearCache();

let currentPath = CUTelectron.app.getAppPath();
let extPath = null;

// Go up until we find an 'web-extensions' sibling
while (currentPath !== CUTpath.dirname(currentPath)) {
    currentPath = CUTpath.dirname(currentPath);
    const testPath = CUTpath.join(currentPath, 'web-extensions');
    if (CUTfs.existsSync(testPath)) {
        extPath = testPath;
        break;
    }
}

// Sentinel reload tracking
let sentinelReloadCount = 0;
const SENTINEL_MAX_RELOADS = 2;
const SENTINEL_TIMEOUT_MS = 5000;
const SENTINEL_STRING = "SENTINEL_EXT_LOADED";
let sentinelReceived = false;

// Load extensions and await them before page navigation
if (extPath) {
    const extDirs = CUTfs.readdirSync(extPath).filter(f =>
        CUTfs.existsSync(CUTpath.join(extPath, f, 'manifest.json'))
    );

    if (extDirs.length > 0) {
        console.log('Loading web extensions...');
        const loadPromises = extDirs.map(f => {
            const p = CUTpath.join(extPath, f);
            console.log('Loading extension:', f);
            return CUTsession.defaultSession.extensions.loadExtension(p).catch(err => {
                console.error('Failed to load extension:', f, err);
            });
        });

        Promise.all(loadPromises).then(() => {
            const loaded = CUTsession.defaultSession.extensions.getAllExtensions().length;
            console.log(`Extensions loaded: ${loaded}/${extDirs.length}`);
            if (loaded < extDirs.length) {
                console.log('Not all extensions loaded, reloading page...');
                sentinelReloadCount++;
                CUTwebView.webContents.reloadIgnoringCache();
            }
        });
    }
}

// Logging + sentinel detection
CUTwebView.webContents.on('console-message', (event) => {
    const message = event.message;
    if (message.startsWith("EXT_LOG:")) {
        console.log(message);
        if (message.includes(SENTINEL_STRING)) {
            sentinelReceived = true;
            console.log('[Sentinel] Content script execution confirmed.');
        }
    }
});

// Sentinel watchdog — check that content scripts executed, retry up to SENTINEL_MAX_RELOADS times
const hasSentinelExtension = extPath && CUTfs.existsSync(CUTpath.join(extPath, 'sentinel', 'manifest.json'));
if (hasSentinelExtension) {
    function checkSentinel() {
        setTimeout(() => {
            if (sentinelReceived) return;
            if (sentinelReloadCount < SENTINEL_MAX_RELOADS) {
                sentinelReloadCount++;
                console.log(`[Sentinel] Content scripts did not execute within ${SENTINEL_TIMEOUT_MS}ms. Reloading (attempt ${sentinelReloadCount}/${SENTINEL_MAX_RELOADS})...`);
                sentinelReceived = false;
                CUTwebView.webContents.reloadIgnoringCache();
                checkSentinel();
            } else {
                console.log(`[Sentinel] Content scripts still not executing after ${SENTINEL_MAX_RELOADS} reloads. Giving up.`);
            }
        }, SENTINEL_TIMEOUT_MS);
    }
    checkSentinel();
}
