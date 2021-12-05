// example.spec.js
const {
    test,
    expect
} = require('@playwright/test');

const {setName, clearRecords} = require('./helpers');

test.describe.parallel('suite', () => {

test('Edit should open/close', async ({ page }) => {
    await setName(page);
    await page.type("[name='A']", '1.2.3.4')
    await page.click('#create')
    const typeBox = page.locator('.desktop .edit')
    await typeBox.click()
    const cancelButton = page.locator('.desktop .cancel')
    await cancelButton.waitFor()
    await typeBox.click()
    await cancelButton.waitFor({state: 'detached'})
    await clearRecords(page)
});

test('Cancel should close the edit form', async ({ page }) => {
    await setName(page);
    await page.type("[name='A']", '1.2.3.4')
    await page.click('#create')
    await page.click('.desktop .edit')
    const cancelButton = page.locator('.desktop .cancel')
    await cancelButton.click()
    await cancelButton.waitFor({state: 'detached'})
    await clearRecords(page)
});

test('Save should update the record and close the form', async ({ page }) => {
    await setName(page);
    await page.type("[name='A']", '1.2.3.4')
    await page.click('#create')
    await page.click('.desktop .edit')
    await page.evaluate(() => {
        document.querySelector("#records [name='A']").value = ''
    })
    await page.type("#records [name='A']", '1.2.3.5')
    await page.click('.save')
    await expect(page.locator('td.view-content')).toHaveText("1.2.3.5 ");
    await clearRecords(page)
});

test('Save should show an error if IP is invalid', async ({ page }) => {
    await setName(page);
    await page.type("[name='A']", '1.2.3.4')
    await page.click('#create')
    await page.click('td .edit')
    await page.evaluate(() => {
        document.querySelector("#records [name='A']").value = ''
    })
    await page.type("#records [name='A']", 'asdf')
    await page.click('.save')
    await expect(page.locator('.formulate-input-error')).toHaveText("Example: 1.2.3.4");
    await clearRecords(page)
});
});
