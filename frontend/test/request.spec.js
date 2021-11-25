// example.spec.js
const {
    test,
    expect
} = require('@playwright/test');

const { randomString, clearRecords } = require('./helpers');

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



const {setName} = require('./helpers');

test.describe.parallel('suite', () => {

test('empty dns request gets streamed', async ({ page }) => {
    const name = randomString();
    await page.goto('http://localhost:8080#' + name);
    const fullName = 'duckface.' + name + '.messwithdns.com.'
    await getIp(fullName);
    await expect(page.locator('.request-name')).toHaveText(fullName);
    await expect(page.locator('.request-host')).toHaveText('localhost.lan.');
    await expect(page.locator('.request-response')).toHaveText('(0 records)');
});

test('clicking request expands it', async ({ page }) => {
    const name = randomString();
    await page.goto('http://localhost:8080#' + name);
    await page.type("[name='A']", '1.2.3.4')
    await page.type("[name='ttl']", '30')
    await page.type("[name='subdomain']", 'bananas');
    await page.click('#create')
    const fullName = 'bananas.' + name + '.messwithdns.com.'
    await getIp(fullName);
    await expect(page.locator('.request-response')).toHaveText('A 1.2.3.4');
    await page.click('.request-response');
    await page.waitForSelector('.expand-request');
    await page.click('.request-response');
    await page.waitForSelector('.expand-request', {state: 'detached'});
});


});
