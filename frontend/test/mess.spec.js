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
  await page.click('#randomSubdomain');

  expect(await page.screenshot()).toMatchSnapshot('random-subdomain.png');
});

test('add and delete A record', async ({ page }) => {
    const browserContext = page.context();
    await browserContext.addInitScript({
        path: 'preload.js'
    });

    await page.goto('http://localhost:8080/#brain-juice');

    await page.waitForSelector('.formulate-input-element--submit--label')

    await page.click('#formulate-global-3')
    await page.click('#formulate-global-4')
    await page.type('#formulate-global-4', '1.2.3.4')
    await page.type('#formulate-global-5', '30')
    await page.click('#formulate-global-6 > .formulate-input-element--submit--label')
    await page.waitForSelector('.edit')
    await page.click('.edit')

    expect(await page.screenshot()).toMatchSnapshot('add-a-record-edit.png');

    await page.waitForSelector('.delete')
    await page.click('.delete')
    await page.on('dialog', dialog => dialog.accept());
    expect(await page.screenshot()).toMatchSnapshot('add-a-record-deleted.png');
});
