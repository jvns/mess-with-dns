// example.spec.js
const { test, expect } = require('@playwright/test');

test('snapshot looks ok', async ({ page }) => {
  const browserContext = page.context();
  await browserContext.addInitScript({
    path: 'preload.js'
  });

  await page.goto('http://localhost:8080');

  expect(await page.screenshot()).toMatchSnapshot('index.png');
});
