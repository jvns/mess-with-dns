// example.spec.js
const {
    test,
    expect
} = require('@playwright/test');

async function loadRecords(page) {
    await page.goto('http://localhost:8080/#toothsome-freezer');
    await page.click('#create');
    await page.click('.view-type');
    await page.click('.request-response');
}

async function randomSeed(page) {
    const browserContext = page.context();
    await browserContext.addInitScript({
        path: 'preload.js'
    });
}

test.describe.parallel('suite', () => {



test('homepage snapshot, desktop', async ({ page }) => {
    await randomSeed(page);
    await page.goto('http://localhost:8080');
    expect(await page.screenshot()).toMatchSnapshot('homepage-desktop.png');
});

test('homepage snapshot, mobile', async ({ page }) => {
    await randomSeed(page);
    await page.setViewportSize({
        width: 320,
        height: 568,
        deviceScaleFactor: 2
    });
    await page.goto('http://localhost:8080');
    expect(await page.screenshot()).toMatchSnapshot('homepage-mobile.png');
});

test('homepage snapshot, tablet', async ({ page }) => {
    await randomSeed(page);
    await page.setViewportSize({
        width: 768,
        height: 1024,
        deviceScaleFactor: 2
    });
    await page.goto('http://localhost:8080');
    expect(await page.screenshot()).toMatchSnapshot('homepage-tablet.png');
});

test('records snapshot, desktop', async ({ page }) => {
    await loadRecords(page);
    await page.setViewportSize({
        width: 1500,
        height: 2000,
        deviceScaleFactor: 2
    });
    expect(await page.screenshot()).toMatchSnapshot('records-desktop.png');
});

test('records snapshot, mobile', async ({ page }) => {
    await loadRecords(page);
    await page.setViewportSize({
        width: 320,
        height: 2500,
        deviceScaleFactor: 2
    });
    // scroll to top
    await page.evaluate(() => {
        window.scrollTo(0, 0);
    });
    expect(await page.screenshot()).toMatchSnapshot('records-mobile.png');
});

test('records snapshot, tablet', async ({ page }) => {
    await loadRecords(page);
    await page.setViewportSize({
        width: 768,
        height: 2024,
        deviceScaleFactor: 2
    });
    expect(await page.screenshot()).toMatchSnapshot('records-tablet.png');
});




});
