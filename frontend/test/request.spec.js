// example.spec.js
const {
    test,
    expect
} = require('@playwright/test');

const { randomString, getSubdomain, setName, goToUsername } = require('./helpers');
const { Resolver } = require('dns');
const resolver = new Resolver();

resolver.setServers(['127.0.0.1:5353']);

function getIp(hostname) {
    return new Promise((resolve, reject) => {
        resolver.resolve4(hostname, (err, addresses) => {
            if (err) {
                resolve(err);
            } else {
                resolve(addresses[0]);
            }
        });
    });
}

test.describe.parallel('suite', () => {

test('empty dns request gets streamed', async ({ page }) => {
    await page.goto('http://localhost:8080');
    await page.click('#start-experimenting');
    const subdomain = await getSubdomain(page);
    const fullName = 'bookface.' + subdomain + '.messwithdns.com.'
    await getIp(fullName);
    page.on('dialog', dialog => dialog.accept());
    await expect(page.locator('.request-name')).toHaveText(fullName);
    await expect(page.locator('.request-host')).toHaveText('localhost.lan.');
    await expect(page.locator('.request-response')).toContainText('Code: NXDOMAIN');
});

test('response content is printed', async ({ page }) => {
    await setName(page, 'bananas');
    const name = await getSubdomain(page);
    await page.type("[name='A']", '1.2.3.4');
    await page.click('#create');
    await page.locator('td.view-name');
    // wait 50ms
    await new Promise(resolve => setTimeout(resolve, 50));
    const fullName = 'bananas.' + name + '.messwithdns.com.';
    await getIp(fullName);
    await expect(page.locator('.request-response')).toContainText('Content: 1.2.3.4')
});

test('clearing requests works', async ({ page }) => {
    await page.goto('http://localhost:8080');
    await page.click('#start-experimenting');
    const subdomain = await getSubdomain(page);
    const fullName = 'bananas.' + subdomain + '.messwithdns.com.'
    await getIp(fullName);
    await expect(page.locator('.request-response')).toContainText('Code: NXDOMAIN');
    page.on('dialog', dialog => dialog.accept());
    await page.click('#clear-requests');
    await page.waitForSelector('.request-response', {state: 'detached'});
});



});
