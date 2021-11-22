// example.spec.js
const {
    test,
    expect
} = require('@playwright/test');

test('homepage snapshot looks ok', async ({
    page
}) => {
    const browserContext = page.context();
    await browserContext.addInitScript({
        path: 'preload.js'
    });

    await page.goto('http://localhost:8080');

    expect(await page.screenshot()).toMatchSnapshot('index.png');
});

test('screenshot of random page looks ok', async ({
    page
}) => {
    const browserContext = page.context();
    await browserContext.addInitScript({
        path: 'preload.js'
    });

    await page.goto('http://localhost:8080/#banana-boat');

    expect(await page.screenshot()).toMatchSnapshot('banana-boat.png');
});

test('subdomain page snapshot', async ({
    page
}) => {
    const browserContext = page.context();
    await browserContext.addInitScript({
        path: 'preload.js'
    });

    await page.goto('http://localhost:8080');
    await page.click('#randomSubdomain');

    expect(await page.screenshot()).toMatchSnapshot('random-subdomain.png');
});

test('add and delete A record', async ({
    page
}) => {
    const browserContext = page.context();
    await browserContext.addInitScript({
        path: 'preload.js'
    });
    page.on('dialog', dialog => dialog.accept());
    await page.goto('http://localhost:8080/#brain-juice');
    await page.waitForSelector('#records')

    await page.type("[name='subname']", "bananas");
    await page.type("[name='A']", '1.2.3.4')
    await page.type("[name='ttl']", '30')
    await page.click('#create')
    await page.waitForSelector('.edit')
    await page.click('.edit')

    expect(await page.screenshot()).toMatchSnapshot('add-a-record-edit.png');

    await page.waitForSelector('.delete')
    await page.click('.delete')
    // sleep for a bit to allow the delete to happen
    await new Promise(resolve => setTimeout(resolve, 500));
    expect(await page.screenshot()).toMatchSnapshot('add-a-record-deleted.png');
});

test('add CNAME record', async ({
    page
}) => {
    const browserContext = page.context();
    await browserContext.addInitScript({
        path: 'preload.js'
    });
    await page.goto('http://localhost:8080/#brain-juice');
    await page.waitForSelector('#records')
    await page.type("[name='subname']", "bananas");
    await page.selectOption("[name='type']", 'CNAME')
    await page.type("[name='target']", 'example.com')
    await page.type("[name='ttl']", '30')
    await page.click('#create')
    expect(await page.screenshot()).toMatchSnapshot('add-cname-record-edit.png');
});

test('test saving record', async ({
    page
}) => {
    const browserContext = page.context();
    await browserContext.addInitScript({
        path: 'preload.js'
    });
    await page.goto('http://localhost:8080/#cord-wrinkle')

    await page.waitForSelector('#records')
    await page.type("[name='subname']", "bananas");
    await page.type("[name='A']", '1.2.3.4')
    await page.type("[name='ttl']", '30')
    await page.click('#create')

    await page.waitForSelector('.edit')
    await page.click('.edit')

    page.on('dialog', dialog => dialog.accept());
    await page.waitForSelector("td [name='ttl']")
    await page.type("td [name='ttl']", '10')
    await page.click('.save')
    await page.waitForSelector("td [name='ttl']", {
        state: 'detached'
    })
    await new Promise(resolve => setTimeout(resolve, 50));
    expect(await page.screenshot()).toMatchSnapshot('save-record.png');
    await page.click('#clear-all')
    await new Promise(resolve => setTimeout(resolve, 100));
});
