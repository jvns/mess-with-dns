export function displayName(domain, record: Record) {
    if (record.subdomain == '@') {
        return domain + ".messwithdns.com";
    } else {
        return record.subdomain + '.' + domain + ".messwithdns.com";
    }
}

export function parseCookies() {
    const cookie = document.cookie;
    const cookies = cookie.split(';');
    const cookieObj = {};
    for (const c of cookies) {
        const cookie = c.split('=');
        cookieObj[cookie[0].trim()] = cookie[1];
    }
    return cookieObj;
}

