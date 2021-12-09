const {
    expect
} = require('@playwright/test');


async function goToUsername(page, name) {
    // get context
    const context = page.context();
    // set cookie
    await context.addCookies([{
        name: 'username',
        value: name,
        url: 'http://localhost:8080',
    }]);
    await page.goto("http://localhost:8080/");
}

function randomString() {
    const length = 16;
    var result = '';
    const characters = 'abcdefghijklmnopqrstuvwxyz';
    const charactersLength = characters.length;
    for (var i = 0; i < length; i++) {
        result += characters.charAt(Math.floor(Math.random() * charactersLength));
    }
    return result;
}

async function setName(page, subdomain) {
    // set name and TTL
    await page.goto('http://localhost:8080');
    await page.click('#start-experimenting');
    await page.waitForSelector("[name='subdomain']")
    await page.type("[name='ttl']", '30')
    if (subdomain === undefined) {
        subdomain = 'test-' + randomString();
    }
    await page.type("[name='subdomain']", subdomain);
    return subdomain;
}

async function createRecord(page, subdomain) {
    await page.click('#create')
    await expect(page.locator('td.view-name')).toContainText(subdomain);
    page.on('dialog', dialog => dialog.accept());
    const delButton = page.locator(".desktop .delete")
    delButton.click()
    // I really don't know why these extra clicks are required,
    // but they seem to make the test less flaky :(
    try {
        await delButton.click({force: true, timeout: 200});
    } catch (e) {
        // ignore
    }
    try {
        await delButton.click({force: true, timeout: 200});
    } catch (e) {
        // ignore
    }
    await page.locator('#records').waitFor({state: 'detached'})
}

async function checkError(page, msg) {
    await page.click('#create')
    await expect(page.locator('.formulate-input-error')).toHaveText(msg);
}

async function clearRecords(page) {
    page.on('dialog', dialog => dialog.accept());
    const clearAll = page.locator('#clear-records');
    await clearAll.click({timeout: 500});
    try {
        await clearAll.click({force: true, timeout: 200});
    } catch (e) {
        // ignore
    }
    try {
        await clearAll.click({force: true, timeout: 200});
    } catch (e) {
        // ignore
    }
}



module.exports = {
    randomString,
    setName,
    createRecord,
    checkError,
    clearRecords,
    goToUsername
}

