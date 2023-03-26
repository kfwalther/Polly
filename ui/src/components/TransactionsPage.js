import { useState, useEffect } from "react";
import { TransactionsTable } from "./TransactionsTable";
import { getDateFromUtcDateTime, toPercent, toUSD } from './Helpers';
import StockLineChart from "./StockLineChart";
import LoadingSpinner from "./LoadingSpinner";

// Helper function to calculate the win rate for our trades (buys/sells).
function getWinRate(txns) {
    if (txns.length > 0) {
        return toPercent((txns.filter(x => x.totalReturn > 0).length / txns.length) * 100.0)
    } else {
        return toPercent(0.0)
    }
}

// Helper function to calculate the market beat rate for our trades (buys/sells).
function getBeatRate(txns) {
    if (txns.length > 0) {
        return toPercent((txns.filter(x => x.excessReturn > 0).length / txns.length) * 100.0)
    } else {
        return toPercent(0.0)
    }
}

// Construct the data table to be displayed on the chart.
function superImposeTradesOnPriceChart(chartData, txnData) {
    var data = []
    var fullSeries = []
    if ((chartData != null) && (chartData.sp500 != null)) {
        // Put the S&P500 values in an array.
        data = chartData.sp500.date.map((k, i) => [
            new Date(k),
            chartData.sp500.close[i],
        ])
        // Iterate through each day market has been open.
        for (let i = 0; i < chartData.sp500.date.length; i++) {
            // Does current date have any txns?
            if (txnData.some(t => getDateFromUtcDateTime(t.dateTime) === getDateFromUtcDateTime(chartData.sp500.date[i]))) {
                // Loop through each txn on this date.
                var txns = txnData.filter(t => getDateFromUtcDateTime(t.dateTime) === getDateFromUtcDateTime(chartData.sp500.date[i]))
                var numBuys = 0
                var numSells = 0
                for (var j = 0; j < txns.length; j++) {
                    // Add a buy event flag positioned above the S&P500 line.
                    if (txns[j].action === "Buy") {
                        numBuys++
                        fullSeries.push([data[i][0], data[i][1], data[i][1] + (numBuys * 5),
                        "Bought " + txns[j].ticker + " @ " + toUSD(txns[j].price) + "\nTotal Return:" + toPercent(txns[j].totalReturn),
                            null, null])
                        // Add a sell event flag positioned below the S&P500 line.
                    } else if (txns[j].action === "Sell") {
                        numSells++
                        fullSeries.push([data[i][0], data[i][1], null, null,
                        data[i][1] - (numSells * 5),
                        "Sold " + txns[j].ticker + " @ " + toUSD(txns[j].price) + "\nTotal Return:" + toPercent(txns[j].totalReturn)])
                    }
                }
            } else {
                fullSeries.push([data[i][0], data[i][1], null, null, null, null])
            }
        }
        // Add the column names at the beginning of the data series.
        fullSeries.unshift(['Date', 'Price', 'Buy', { role: 'tooltip' }, 'Sell', { role: 'tooltip' }])
    }
    return fullSeries
}

// Fetch the transaction data from the server, and return/render the transactions page.
export default function TransactionsPage() {
    const [isLoading, setIsLoading] = useState(false);
    const [buySellList, setBuySellList] = useState([]);
    const [chartDataSeries, setChartDataSeries] = useState([]);
    const [errorMessage, setErrorMessage] = useState("");

    document.body.style.backgroundColor = "black"
    // A flag to set so we ignore the second useEffect call, so data isn't fetched twice.
    let ignoreEffect = false;

    // Simple function to perform async fetch of transaction data.
    function getTransactions() {
        return fetch('http://localhost:5000/transactions')
            .then(resp => resp.json())
            .then(json => json["transactions"])
    }

    // Simple function to perform async fetch of S&P500 data.
    function getSp500Data() {
        return fetch('http://localhost:5000/sp500')
            .then(resp => resp.json())
    }

    // A function to setup a Promise to synchronously wait for all fetches to finish.
    function getAllData() {
        return Promise.all([getTransactions(), getSp500Data()])
    }

    // Runs on mount (twice, by design), so use ignore flag so we don't fetch and filter twice.
    useEffect(() => {
        // Indicate we're in loading state, so spinner is displayed.
        setIsLoading(true);
        console.log('Fetching transaction data...')

        // Ignore when this useEffect is called the second time.
        if (!ignoreEffect) {
            // Fetch all the data and wait for it to populate here.
            getAllData()
                .then(([txns, sp500Data]) => {
                    console.log('Done fetching! Txns: ' + txns.length + ', SP500 Data: ' + sp500Data.sp500.close.length)
                    console.log('Filtering data...')
                    if (txns != null && txns.length > 0) {
                        // Filter txns for only buy/sell actions.
                        var filteredList = txns.filter(t => (t.action === "Buy" || t.action === "Sell"))
                        setBuySellList(filteredList)
                        // Generate the data series with trades super-imposed on the S&P500 price data.
                        var series = superImposeTradesOnPriceChart(sp500Data, filteredList)
                        setChartDataSeries(series)
                        console.log('Done filtering!')
                    }
                }
                )
        }

        // Cleanup function, so double useEffect doesn't keep affecting our fetched data.
        return () => {
            ignoreEffect = true;
            console.log('Cleaned!')
        };
    }, []);

    // Callback that runs after final chart series data is updated with useState hook.
    useEffect(() => {
        if (chartDataSeries.length > 0) {
            console.log('Series has been fully populated!')
            // Indicate loading has finished.
            setIsLoading(false);
        }
    }, [chartDataSeries]);

    // Return this JSX content to be rendered.
    return (
        <>
            {/* Display a rollup/summary of all our transactions. */}
            <table className="txn-summary-table">
                <th className="txn-summary-table-header">Transaction Count</th>
                <th className="txn-summary-table-header">Win Rate</th>
                <th className="txn-summary-table-header">Beat Rate</th>
                <tbody>
                    <tr>
                        <td className="txn-summary-table-cell">{buySellList.length}</td>
                        <td className="txn-summary-table-cell">{getWinRate(buySellList)}</td>
                        <td className="txn-summary-table-cell">{getBeatRate(buySellList)}</td>
                    </tr>
                </tbody>
            </table>
            {/* Add a line chart to superimpose our trades on the S&P500 performance. */}
            {(isLoading === true || chartDataSeries.length === 0) ? <LoadingSpinner /> :
                <StockLineChart
                    chartDataSeries={chartDataSeries}
                />
            }
            {errorMessage && <div className="error">{errorMessage}</div>}
            <h3 className="header-centered">Transactions List</h3>
            {/* Display all the transactions in a sortable table. */}
            {(isLoading === true || buySellList.length === 0) ? <LoadingSpinner /> :
                <TransactionsTable
                    txnData={buySellList}
                />
            }
        </>
    );
}