
// A helper function to format a string with a specific number of decimal places
// export function withDecimalPlaces(num) {
//     return parseFloat(numberString)
// }

const usdFormatter = new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'USD',
});

// Define nested function to format the raw data from the server to USD format.
export function toUSD(numberString) {
    let number = parseFloat(numberString);
    return usdFormatter.format(number);
}

export function toPercent(numberString) {
    let number = parseFloat(numberString);
    return number.toFixed(2) + "%";
}

// Extract the date portion from the front of an ISO-formatted UTC datetime string.
export function getDateFromUtcDateTime(dateTime) {
    return dateTime.substring(0, dateTime.indexOf("T"));
}
