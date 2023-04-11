import { useState, useEffect } from "react";
import { TransactionsTable } from "./TransactionsTable";
import StockLineChart from "./StockLineChart";
import LoadingSpinner from "./LoadingSpinner";

// Construct the data table to be displayed on the chart.
function calculatePortfolioHistory(historyData) {
    var r = historyData.filter(s => s.ticker == "TSLA")[0].valueHistory
    var my = []
    for (const [date, value] of Object.entries(r)) {
        my.push([new Date(date), value])
    }
    console.log(my)
    // Add the column names at the beginning of the data series.
    my.unshift(['Date', 'Price'])

    return my
}

// Fetch the transaction data from the server, and return/render the transactions page.
export default function HistoryPage() {
    const [isLoading, setIsLoading] = useState(false);
    const [buySellList, setBuySellList] = useState([]);
    const [chartDataSeries, setChartDataSeries] = useState([]);
    const [errorMessage, setErrorMessage] = useState("");

    document.body.style.backgroundColor = "black"
    // A flag to set so we ignore the second useEffect call, so data isn't fetched twice.
    let ignoreEffect = false;

    // Simple function to perform async fetch of history data.
    function getSecurities() {
        return fetch('http://localhost:5000/securities')
            .then(resp => resp.json())
            .then(json => json["securities"])
    }

    function getHistoryData() {
        return fetch("http://localhost:5000/history")
            .then(resp => resp.json())
            .then(json => json["history"])
    }

    // A function to setup a Promise to synchronously wait for all fetches to finish.
    function getAllData() {
        return Promise.all([getSecurities(), getHistoryData()])
    }

    // Runs on mount (twice, by design), so use ignore flag so we don't fetch and filter twice.
    useEffect(() => {
        // Indicate we're in loading state, so spinner is displayed.
        setIsLoading(true);
        console.log('Fetching history data...')

        // Ignore when this useEffect is called the second time.
        if (!ignoreEffect) {
            // Fetch all the data and wait for it to populate here.
            getAllData()
                .then(([stockData, historyData]) => {
                    console.log('Done fetching! Stock data: ' + stockData.length)
                    console.log('History data: ' + historyData)
                    console.log('Calculating portfolio totals...')
                    if (stockData != null && stockData.length > 0) {
                        // Add up all the stock totals for each day.
                        // TODO: FIX PLOTTING!
                        // var series = calculatePortfolioHistory(historyData)
                        // setChartDataSeries(series)
                        console.log('Done calculating!')
                    }
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

    // Return this JSX content to be rendered.
    return (
        <>
            {/* Display the portfolio history chart. */}
            {(isLoading === true || chartDataSeries.length === 0) ? <LoadingSpinner /> :
                <StockLineChart
                    chartDataSeries={chartDataSeries}
                />
            }
            {errorMessage && <div className="error">{errorMessage}</div>}
        </>
    );
}