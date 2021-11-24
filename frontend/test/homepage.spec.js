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
    const text = await page.innerText('.add-record');
    expect(text).toBe('Add a record');
});

test('typing in domain brings us to domain page', async ({
    page
}) => {
    await page.goto('http://localhost:8080');
    await page.type("[name='domain']", "fruity-pebble");
    await page.click("[type='submit']");
    const text = await page.innerText('#subdomain-title');
    expect(text).toBe('Your subdomain is:\nfruity-pebble.messwithdns.com')
});
