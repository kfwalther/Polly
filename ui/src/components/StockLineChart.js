import { Chart } from 'react-google-charts';
import { getDateFromUtcDateTime, toPercent, toUSD } from './Helpers';


export default function StockLineChart({ chartData, txnData }) {

    // Construct the data table to be displayed on the chart.
    function getFilteredChartData() {
        var data = []
        var fullSeries = []
        if ((chartData != null) && (chartData.sp500 != null)) {
            console.log(chartData.sp500)
            console.log(txnData)
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

    // Define the options for this line chart.
    var chartOptions = {
        legend: 'none',
        backgroundColor: 'transparent',
        title: 'S&P500 Performance',
        titleTextStyle: {
            color: 'lightgrey'
        },
        colors: ['aqua'],
        crosshair: { orientation: 'vertical', trigger: 'focus', color: 'lightgrey' },
        curveType: 'function',
        chartArea: { top: 25, bottom: 50, left: 50, right: 25 },
        hAxis: {
            format: 'MMM y',
            gridlines: { count: 5, color: 'transparent' },
            minorGridlines: { count: 0 },
        },
        vAxis: {
            gridlines: { count: 0 },
            minorGridlines: { count: 0 },
            viewWindowMode: 'maximized',
        },
        series: {
            1: {
                type: 'scatter',
                color: 'green',
                visibleInLegend: false
            },
            2: {
                type: 'scatter',
                color: 'red',
                visibleInLegend: false
            },
        },
        responsive: true,
    }
    // Format the incoming data to be displayed in the line chart.
    var data = getFilteredChartData()
    // Define the height of the chart range filter so we can pad for it.
    var filterHeight = 50

    return (
        // Apply bottom margin around this div container to account for the height of chart range filter.
        <div className="txnchart-container" style={{ marginBottom: filterHeight + 25 }}>
            <Chart
                chartType="LineChart"
                width="100%"
                height="600px"
                data={data}
                options={chartOptions}
                controls={[
                    {
                        controlType: "ChartRangeFilter",
                        options: {
                            filterColumnIndex: 0,
                            ui: {
                                // Line chart for the range filter display also!
                                chartType: "LineChart",
                                chartOptions: {
                                    // Set height of filter and height of chart within filter obj to be the same.
                                    height: filterHeight,
                                    chartArea: {
                                        width: "95%",
                                        height: filterHeight,
                                    },
                                    backgroundColor: "black",
                                    hAxis: {
                                        gridlines: { count: 4 },
                                        minorGridlines: { count: 0 },
                                    },
                                },
                            },
                        },
                        controlPosition: "bottom",
                        controlWrapperParams: {
                            state: {
                                range: {
                                    // Set the range filter at beginning of 2020 (zero-indexed for month/day).
                                    start: new Date(2020, 0, 0),
                                },
                            },
                        },
                    },
                ]}
            />
        </div >
    );
}