// example.spec.js
const {
    test,
    expect
} = require('@playwright/test');

const {randomString, setName, createRecord, checkError, getSubdomain} = require('./helpers');


test.describe.parallel('suite', () => {

test('A record should appear when added', async ({ page }) => {
    const subdomain = await setName(page);
    await page.type("[name='A']", '1.2.3.4')
    await createRecord(page, subdomain);
});

test('AAAA record should appear when added', async ({ page }) => {
    const subdomain = await setName(page);
    await page.selectOption("[name='type']", 'AAAA')
    await page.type("[name='AAAA']", '2607:f8b0:4020:807::200e')
    await createRecord(page, subdomain);
});

test('CNAME record should appear when added', async ({ page }) => {
    const subdomain = await setName(page);
    await page.selectOption("[name='type']", 'CNAME')
    await page.type("[name='Target']", 'example.com')
    await createRecord(page, subdomain);
});

test('CAA record should appear when added', async ({ page }) => {
    const subdomain = await setName(page);
    await page.selectOption("[name='type']", 'CAA')
    await page.type("[name='Flag']", '30')
    await page.type("[name='Tag']", '30')
    await page.type("[name='Value']", 'example.com')
    await createRecord(page, subdomain);
});

test('CERT record should appear when added', async ({ page }) => {
    const subdomain = await setName(page);
    await page.selectOption("[name='type']", 'CERT')
    await page.type("[name='KeyTag']", '30')
    await page.type("[name='Type']", '30')
    await page.type("[name='Algorithm']", '30')
    await page.type("[name='Certificate']", 'asdf')
    await createRecord(page, subdomain);
});

test('MX record should appear when added', async ({ page }) => {
    const subdomain = await setName(page);
    await page.selectOption("[name='type']", 'MX')
    await page.type("[name='Preference']", '10')
    await page.type("[name='Mx']", 'mail.example.com')
    await createRecord(page, subdomain);
});

test('NS record should appear when added', async ({ page }) => {
    const subdomain = await setName(page);
    await page.selectOption("[name='type']", 'NS')
    await page.type("[name='Ns']", 'ns1.example.com')
    await createRecord(page, subdomain);
});


test('PTR record should appear when added', async ({ page }) => {
    const subdomain = await setName(page);
    await page.selectOption("[name='type']", 'PTR')
    await page.type("[name='Ptr']", 'example.com')
    await createRecord(page, subdomain);
});

test('TXT record should appear when added', async ({ page }) => {
    const subdomain = await setName(page);
    await page.selectOption("[name='type']", 'TXT')
    await page.type("[name='Txt']", 'asdf')
    await createRecord(page, subdomain);
});

test('URI record should appear when added', async ({ page }) => {
    const subdomain = await setName(page);
    await page.selectOption("[name='type']", 'URI')
    await page.type("[name='Priority']", '30')
    await page.type("[name='Weight']", '30')
    await page.type("[name='Target']", 'example.com')
    await createRecord(page, subdomain);
});

test('SRV record should appear when added', async ({ page }) => {
    const subdomain = await setName(page);
    await page.selectOption("[name='type']", 'SRV')
    await page.type("[name='Priority']", '30')
    await page.type("[name='Weight']", '30')
    await page.type("[name='Port']", '8080')
    await page.type("[name='Target']", 'example.com')
    await createRecord(page, subdomain);
});

test('DS record should appear when added', async ({ page }) => {
    const subdomain = await setName(page);
    await page.selectOption("[name='type']", 'DS')
    await page.type("[name='KeyTag']", '30')
    await page.type("[name='Algorithm']", '30')
    await page.type("[name='DigestType']", '30')
    await page.type("[name='Digest']", 'affe2345')
    await createRecord(page, subdomain);
});

test('SOA record should appear when added', async ({ page }) => {
    const subdomain = await setName(page);
    await page.selectOption("[name='type']", 'SOA')
    await page.type("[name='Ns']", 'ns1.example.com')
    await page.type("[name='Mbox']", 'hostmaster.example.com')
    await page.type("[name='Serial']", '30')
    await page.type("[name='Refresh']", '30')
    await page.type("[name='Retry']", '30')
    await page.type("[name='Expire']", '30')
    await page.type("[name='Minttl']", '30')
    await createRecord(page, subdomain);
});

test('@ record works', async ({ page }) => {
    await setName(page, '@');
    await page.type("[name='A']", '1.2.3.4')
    await page.click('#create')
    const subdomain = await getSubdomain(page);
    await expect(page.locator('.desktop .view-name')).toHaveText(subdomain + ".messwithdns.net");
    page.on('dialog', dialog => dialog.accept());
    const delButton = page.locator(".desktop .delete")
    delButton.click()
    // I really don't know why these extra clicks are required,
    // but they seem to make the test less flaky :(
    delButton.click({force: true})
    await page.locator('#records').waitFor({state: 'detached'})
})



/********************
 * error message tests
 ******************/

test('A record error message', async ({ page }) => {
    await setName(page);
    await page.type("[name='A']", 'asdf')
    await checkError(page, 'Example: 1.2.3.4')
})

test('AAAA record error message', async ({ page }) => {
    await setName(page);
    await page.selectOption("[name='type']", 'AAAA')
    await page.type("[name='AAAA']", 'xxx')
    await checkError(page, 'Example: 2001:db8::1')
})

test('CNAME record error message', async ({ page }) => {
    await setName(page);
    await page.selectOption("[name='type']", 'CNAME')
    await page.type("[name='Target']", 'asdf')
    await checkError(page, 'Example: orange-ip.fly.dev')
})

test('ttl error message', async ({ page }) => {
    await page.goto('http://localhost:8080');
    await page.click('#start-experimenting');
    const subdomain = 'test-' + randomString();
    await page.type("[name='subdomain']", subdomain);
    await page.type("[name='A']", '1.2.3.4')
    await checkError(page, 'Example: 60')
})

test("name can't be blank", async ({ page }) => {
    await page.goto('http://localhost:8080');
    await page.click('#start-experimenting');
    await page.type("[name='A']", '1.2.3.4')
    await page.type("[name='ttl']", '30')
    await checkError(page, "Example: bananas")
})

test('server error message', async ({ page }) => {
    await setName(page);
    await page.selectOption("[name='type']", 'AAAA')
    await page.type("[name='AAAA']", 'asdf')
    await page.click('#create')
    await expect(page.locator('.server-error')).toHaveText("Error parsing record: invalid IP address: asdf")
})

test('server error message: bad domain name', async ({ page }) => {
    await page.goto('http://localhost:8080');
    await page.click('#start-experimenting');
    await page.waitForSelector("[name='subdomain']")
    await page.type("[name='ttl']", '30')

    await page.type("[name='subdomain']", "banana..asdf");
    await page.type("[name='A']", '1.2.3.4')
    await page.click('#create')
    await expect(page.locator('.server-error')).toHaveText("Oops, invalid domain name")
})


});
