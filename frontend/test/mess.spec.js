// example.spec.js
const { test, expect } = require('@playwright/test');

test('homepage snapshot looks ok', async ({ page }) => {
  const browserContext = page.context();
  await browserContext.addInitScript({
    path: 'preload.js'
  });

  await page.goto('http://localhost:8080');

  expect(await page.screenshot()).toMatchSnapshot('index.png');
});

test('screenshot of random page looks ok', async ({ page }) => {
  const browserContext = page.context();
  await browserContext.addInitScript({
    path: 'preload.js'
  });

  await page.goto('http://localhost:8080/#banana-boat');

  expect(await page.screenshot()).toMatchSnapshot('banana-boat.png');
});

test('subdomain page snapshot', async ({ page }) => {
  const browserContext = page.context();
  await browserContext.addInitScript({
    path: 'preload.js'
  });

  await page.goto('http://localhost:8080');
  await page.click('#randomSubdomain', {timeout: 1000});

  expect(await page.screenshot()).toMatchSnapshot('random-subdomain.png');
});
