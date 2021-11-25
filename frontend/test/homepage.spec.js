// example.spec.js
const {
    test,
    expect
} = require('@playwright/test');

test('clicking random subdomain brings us to subdomain page', async ({
    page
}) => {
    await page.goto('http://localhost:8080');
    await page.click('#randomSubdomain');
    await expect(page.locator('.add-record')).toHaveText('Add a record');
});

test('typing in domain brings us to domain page', async ({
    page
}) => {
    await page.goto('http://localhost:8080');
    await page.type("[name='domain']", "fruity-pebble");
    await page.click("[type='submit']");
    await expect(page.locator('#subdomain-title')).toHaveText('Your subdomain is:\nfruity-pebble.messwithdns.com');
});
