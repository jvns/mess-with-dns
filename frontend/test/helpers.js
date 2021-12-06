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

async function setName(page) {
    // set name and TTL
    await page.goto('http://localhost:8080');
    await page.click('#randomSubdomain');
    await page.waitForSelector("[name='subdomain']")
    await page.type("[name='ttl']", '30')
    const subdomain = 'test-' + randomString();
    await page.type("[name='subdomain']", subdomain);
    return subdomain;
}

async function createRecord(page, subdomain) {
    await page.click('#create')
    await expect(page.locator('td.view-name')).toContainText(subdomain);
    page.on('dialog', dialog => dialog.accept());
    await page.click('.edit')
    // I don't know why, but this test is flaky if we only delete once :(
    await page.waitForSelector(".delete")
    const delButton = page.locator(".delete")
    delButton.click()
    // I really don't know why these extra clicks are required,
    // but they seem to make the test less flaky :(
    delButton.click({force: true})
    delButton.click({force: true})
    await page.locator('#records').waitFor({state: 'detached'})
}

async function checkError(page, msg) {
    await page.click('#create')
    await expect(page.locator('.formulate-input-error')).toHaveText(msg);
}

async function clearRecords(page) {
    page.on('dialog', dialog => dialog.accept());
    const clearAll = page.locator('#clear-records');
    await clearAll.click();
    await clearAll.click({force: true});
    await clearAll.click({force: true});
}



module.exports = {
    randomString,
    setName,
    createRecord,
    checkError,
    clearRecords,
    goToUsername
}

