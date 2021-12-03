// example.spec.js
const {
    test,
    expect
} = require('@playwright/test');

test.describe.parallel('suite', () => {

test('all the modals open', async ({ page }) => {
    // wait for everything to be mounted
    await page.goto('http://localhost:8080#reptile-brain');
    await page.waitForSelector('.details-modal-close', {state: 'attached'});
    // select all the modals
    const experiments = await page.$$('.experiment');
    // open and close all the modals sequentially
    for (let i = 0; i < experiments.length; i++) {
        const exp = experiments[i];
        const detailsModal = await exp.$('.details-modal');
        expect(await detailsModal.isVisible()).toBe(false);
        const summary = await exp.$('summary');
        await summary.click();
        expect(await detailsModal.isVisible()).toBe(true);
        // get .details-modal child
        // check if it's visible
        // close the modal
        const overlay = await exp.$('.details-modal-close');
        await overlay.click({force: true});
        // check that it's not visible
        expect(await detailsModal.isVisible()).toBe(false);
    }
    /*
    await Promise.all(experiments.map(async exp => {
    }));
    */
});



});
