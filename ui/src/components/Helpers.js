
// A helper function to format a string with a specific number of decimal places
// export function withDecimalPlaces(num) {
//     return parseFloat(numberString)
// }

// Define nested function to format the raw data from the server to USD format.
export default function toUSD(numberString) {
    let number = parseFloat(numberString);
    return "$" + number.toFixed(2);
}
