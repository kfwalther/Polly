
// A helper function to format a string with a specific number of decimal places
// export function withDecimalPlaces(num) {
//     return parseFloat(numberString)
// }

// Define nested function to format the raw data from the server to USD format.
export function toUSD(numberString) {
    let number = parseFloat(numberString);
    let isNeg = number < 0.0 ? '-' : ''
    return isNeg + '$' + Math.abs(number).toFixed(2);
}

export function toPercent(numberString) {
    let number = parseFloat(numberString);
    return number.toFixed(2) + "%";
}

// Extract the date portion from the front of an ISO-formatted UTC datetime string.
export function getDateFromUtcDateTime(dateTime) {
    return dateTime.substring(0, dateTime.indexOf("T"));
}
