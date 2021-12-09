// example.spec.js
const {
    test,
    expect
} = require('@playwright/test');

test('clicking random subdomain brings us to subdomain page', async ({
    page
}) => {
    await page.goto('http://localhost:8080');
    await page.click('#start-experimenting');
    await expect(page.locator('.add-record')).toHaveText('Add a record');
});

