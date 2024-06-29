import rrTypes from './rrTypes.json';
import schemas from './schemas.json';
const punycode = require('punycode');

export const rrTypesReverse = {};
for (const key in rrTypes) {
    rrTypesReverse[rrTypes[key]] = key;
}

export function displayName(record: Record) {
    if (record.subdomain == '@') {
        return record.domain + ".messwithdns.com";
    } else {
        return record.subdomain + '.' + record.domain + ".messwithdns.com";
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

