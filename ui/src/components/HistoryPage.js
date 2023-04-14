import { useState, useEffect } from "react";
import Select from 'react-select';
import StockLineChart from "./StockLineChart";
import LoadingSpinner from "./LoadingSpinner";
import "./HistoryPage.css";

// Construct the data table to be displayed on the chart.
function convertValueHistoryToSeries(historyData) {
    var series = []
    // Reformat object of date-price key-value pairs into series for plotting.
    for (const [date, value] of Object.entries(historyData)) {
        series.push([new Date(date), value])
    }
    // Add the column names at the beginning of the data series.
    series.unshift(['Date', 'Price'])
    return series
}

// Fetch the transaction data from the server, and return/render the transactions page.
export default function HistoryPage() {
    const [isLoading, setIsLoading] = useState(false);
    const [stockData, setStockData] = useState([]);
    const [stockList, setStockList] = useState([]);
    const [chartDataSeries, setChartDataSeries] = useState([]);

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
                .then(([stocks, historyData]) => {
                    console.log('Done fetching! Number of stocks retrieved: ' + stocks.length)
                    // If we got the list of stocks, filter out CASH position, and save them.
                    if (stocks != null && stocks.length > 0) {
                        var onlyStocks = stocks.filter(s => s.ticker !== "CASH")
                        setStockData(onlyStocks)
                        var simpleList = onlyStocks.map(s => ({ value: s.ticker, label: s.ticker }))
                        setStockList(simpleList)
                    }
                    // If we got total portfolio history, convert it to a plot series, and save.
                    if (historyData != null) {
                        var series = convertValueHistoryToSeries(historyData)
                        setChartDataSeries(series)
                        console.log('Done formatting plot series data.')
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

    // Plot the selected stock in the line chart.
    function plotSelectedStock(selection) {
        setIsLoading(true);
        console.log('Plotting ' + selection.value + '...')
        var dataToPlot = stockData.find(s => s.ticker === selection.value).valueHistory
        var series = convertValueHistoryToSeries(dataToPlot)
        // When we save here, the useEffect will run and refresh the page.
        setChartDataSeries(series)
    }

    // Return this JSX content to be rendered.
    return (
        <>
            {/* Display the portfolio history chart. */}
            {(isLoading === true || chartDataSeries.length === 0) ? <LoadingSpinner /> :
                <StockLineChart
                    chartDataSeries={chartDataSeries}
                    chartTitle={'Historical Performance'}
                />
            }
            {/* Show a drop-down list to allow user to specify which stock to plot. */}
            <h3 className="stock-picker-label">Pick a ticker to plot:</h3>
            <div className="stock-picker-container">
                <Select options={stockList} onChange={plotSelectedStock} />
            </div>
        </>
    );
}