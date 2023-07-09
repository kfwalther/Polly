import { useState, useEffect } from "react";
import Select from 'react-select';
import { TransactionsTable } from "./TransactionsTable";
import { getDateFromUtcDateTime, toPercent, toUSD } from './Helpers';
import StockLineChart from "./StockLineChart";
import LoadingSpinner from "./LoadingSpinner";
import "./TransactionsPage.css";

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
function convertValueHistoryToSeries(historyData) {
    var series = []
    // Reformat object of date-price key-value pairs into series for plotting.
    for (const [date, value] of Object.entries(historyData)) {
        series.push([new Date(date), value])
    }
    return series
}

// Find the max of the stock data value (y-axis or second column).
function findMax(stockData) {
    const max = stockData.reduce(([max], [_, second]) => [
        Math.max(max, second)
    ], [stockData[0][1]]);
    return max
}

// Find the min/max of the stock data value (y-axis or second column).
function findMinMax(stockData) {
    const [min, max] = stockData.reduce(([min, max], [_, second]) => [
        Math.min(min, second),
        Math.max(max, second)
    ], [stockData[0][1], stockData[0][1]]);
    return [min, max]
}

// Construct the data table to be displayed on the chart.
function superImposeTradesOnPriceChart(stockData, txnData) {
    var fullSeries = []
    if (stockData != null && stockData.length > 0 && txnData != null && txnData.length > 0) {
        // Determine the min and max values for the y-axis, so we can properly space the txn flags when plotted.
        const [min, max] = findMinMax(stockData)
        // Calculate the txn flag spacing
        var flagSpacing = (max - min) * 0.03
        // Iterate through each day market has been open.
        for (let i = 0; i < stockData.length; i++) {
            // Does current date have any txns?
            if (txnData.some(t => getDateFromUtcDateTime(t.dateTime) === getDateFromUtcDateTime(stockData[i][0].toISOString()))) {
                // Loop through each txn on this date.
                var txns = txnData.filter(t => getDateFromUtcDateTime(t.dateTime) === getDateFromUtcDateTime(stockData[i][0].toISOString()))
                var numBuys = 0
                var numSells = 0
                for (var j = 0; j < txns.length; j++) {
                    // Add a buy event flag positioned above the plotted line.
                    if (txns[j].action === "Buy") {
                        numBuys++
                        fullSeries.push([stockData[i][0], stockData[i][1], stockData[i][1] + (numBuys * flagSpacing),
                        "Bought " + txns[j].ticker + " @ " + toUSD(txns[j].price) + "\nValue: " + toUSD(txns[j].value) + "\nTotal Return: " + toPercent(txns[j].totalReturn),
                            null, null])
                        // Add a sell event flag positioned below the plotted line.
                    } else if (txns[j].action === "Sell") {
                        numSells++
                        fullSeries.push([stockData[i][0], stockData[i][1], null, null,
                        stockData[i][1] - (numSells * flagSpacing),
                        "Sold " + txns[j].ticker + " @ " + toUSD(txns[j].price) + "\nValue: " + toUSD(txns[j].value) + "\nTotal Return: " + toPercent(txns[j].totalReturn)])
                    }
                }
            } else {
                fullSeries.push([stockData[i][0], stockData[i][1], null, null, null, null])
            }
        }
        // Add the column names at the beginning of the data series.
        fullSeries.unshift([{ label: 'Date', type: 'date' }, { label: 'Price', type: 'number' },
        { label: 'Buy', type: 'number' }, { role: 'tooltip' }, { label: 'Sell', type: 'number' }, { role: 'tooltip' }])
    }
    return fullSeries
}

// Fetch the transaction data from the server, and return/render the transactions page.
export default function TransactionsPage() {
    // Define an isLoading flag.
    const [isLoading, setIsLoading] = useState(false);
    // Define a plot title descriptor.
    const [plotDesc, setPlotDesc] = useState('Total Portfolio')
    // Define the current and max value plotted on the chart.
    const [curValue, setCurValue] = useState(0.0)
    const [maxValue, setMaxValue] = useState(0.0)
    // This is where we store the complete txns list of buys and sells.
    const [buySellList, setBuySellList] = useState([]);
    // This is the currently displayed txn data for the table.
    const [txnTableData, setTxnTableData] = useState([]);
    // This is the full list of Security objects from the backend.
    const [stockData, setStockData] = useState([]);
    // A list of stock tickers to display in the drop-down list.
    const [stockTickerList, setStockTickerList] = useState([]);
    // The current data series plotted on the graph.
    const [chartDataSeries, setChartDataSeries] = useState([]);

    document.body.style.backgroundColor = "black"
    // A flag to set so we ignore the second useEffect call, so data isn't fetched twice.
    let ignoreEffect = false;

    // Simple function to perform async fetch of transaction data.
    function getTransactions() {
        return fetch("http://" + process.env.REACT_APP_API_BASE_URL + "/transactions")
            .then(resp => resp.json())
            .then(json => json["transactions"])
    }

    // Simple function to perform async fetch of history data.
    function getSecurities() {
        return fetch("http://" + process.env.REACT_APP_API_BASE_URL + "/securities")
            .then(resp => resp.json())
            .then(json => json["securities"])
    }

    function getHistoryData() {
        return fetch("http://" + process.env.REACT_APP_API_BASE_URL + "/history")
            .then(resp => resp.json())
            .then(json => json["history"])
    }

    // A function to setup a Promise to synchronously wait for all fetches to finish.
    function getAllData() {
        return Promise.all([getTransactions(), getSecurities(), getHistoryData()])
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
                .then(([txns, stocks, historyData]) => {
                    console.log('Done fetching! Txns: ' + txns.length + ', Stock Count: ' + stocks.length)
                    console.log('Filtering data...')
                    var filteredTxnList = []
                    var historySeries = []
                    if (txns != null && txns.length > 0) {
                        // Filter txns for only buy/sell actions.
                        filteredTxnList = txns.filter(t => (t.action === "Buy" || t.action === "Sell"))
                        setBuySellList(filteredTxnList)
                        setTxnTableData(filteredTxnList)
                    }
                    // If we got the list of stocks, filter out CASH position, and save them.
                    if (stocks != null && stocks.length > 0) {
                        var onlyStocks = stocks.filter(s => s.ticker !== "CASH")
                        setStockData(onlyStocks)
                        // Get a list of tickers, sorted alphabetically, for the drop-down list labels.
                        var tickerList = onlyStocks.map(s => ({ value: s.ticker, label: s.ticker }))
                        tickerList.sort((a, b) => a.label.localeCompare(b.label))
                        setStockTickerList(tickerList)
                    }
                    // If we got total portfolio history, convert it to a plot series, and save.
                    if (historyData != null) {
                        historySeries = convertValueHistoryToSeries(historyData)
                        // Get the current and all-time high portfolio value.
                        setCurValue(historySeries[historySeries.length - 1][1])
                        setMaxValue(findMax(historySeries))
                    }
                    // Generate the data series with trades super-imposed on the historical price data.
                    var series = superImposeTradesOnPriceChart(historySeries, filteredTxnList)
                    setChartDataSeries(series)
                    console.log('Done filtering!')
                })
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

    // Plot the selected stock in the line chart.
    function plotSelectedStock(selection) {
        setIsLoading(true);
        console.log('Plotting ' + selection.value + '...')
        // Update the title descriptor of the plot.
        setPlotDesc(selection.value)
        // Filter for only the specific stock's data, then convert it to a plottable series.
        var stockToPlot = stockData.find(s => s.ticker === selection.value)
        var stockSeries = convertValueHistoryToSeries(stockToPlot.valueHistory)
        // Save the current and all-time high value.
        setCurValue(stockToPlot.marketValue)
        setMaxValue(stockToPlot.valueAllTimeHigh)
        // Filter the transactions for the specific stock's data.
        var txnsToOverlay = buySellList.filter(t => (t.ticker === selection.value))
        // Update the txn table with only these txns.
        setTxnTableData(txnsToOverlay)
        // Generate a full series of data with transactions overlayed on the stock data.
        var series = superImposeTradesOnPriceChart(stockSeries, txnsToOverlay)
        // When we save here, the useEffect will run and refresh the page.
        setChartDataSeries(series)
    }

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
            {/* Add some metrics below the summary data. */}
            <span className="metrics-span">{'Current Value: ' + toUSD(curValue) + ' | All-Time High: ' + toUSD(maxValue)}</span>
            {/* Add a line chart to superimpose our trades on the portfolio's historical performance. */}
            {(isLoading === true || chartDataSeries.length === 0) ? <LoadingSpinner /> :
                <StockLineChart
                    chartDataSeries={chartDataSeries}
                    chartTitle={'Historical Performance for ' + plotDesc}
                />
            }
            {/* Show a drop-down list to allow user to specify which stock to plot. */}
            <h3 className="stock-picker-label">Pick a ticker to plot:</h3>
            <div className="stock-picker-container">
                <Select options={stockTickerList} onChange={plotSelectedStock} />
            </div>
            <h3 className="header-centered">Transactions List</h3>
            {/* Display all the transactions in a sortable table. */}
            {(isLoading === true || txnTableData.length === 0) ? <LoadingSpinner /> :
                <TransactionsTable
                    txnData={txnTableData}
                />
            }
        </>
    );
}