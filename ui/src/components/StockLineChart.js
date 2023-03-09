import { Chart } from 'react-google-charts';
import { getDateFromUtcDateTime } from './Helpers';


export default function StockLineChart({ chartData, txnData }) {

    function getFilteredChartData() {
        var data = []
        if ((chartData != null) && (chartData.sp500 != null)) {
            console.log(chartData.sp500)
            console.log(txnData)
            // Put the sorted values in an array, and add a column header.
            data = chartData.sp500.date.map((k, i) => [
                new Date(k),
                chartData.sp500.close[i],
                (txnData.some(t => getDateFromUtcDateTime(t.dateTime) === getDateFromUtcDateTime(k) &&
                        t.action === "Buy") ? "B" : null),
                (txnData.some(t => getDateFromUtcDateTime(t.dateTime) === getDateFromUtcDateTime(k) &&
                        t.action === "Sell") ? "S" : null),
            ])
            data.unshift(['Date', 'Price', {role: 'annotation'}, {role: 'annotation'}])
        }
        return data
    }

    // Define the options for this line chart.
    var chartOptions = {
        legend: 'none',
        backgroundColor: 'transparent',
        title: 'S&P500 Performance',
        titleTextStyle: {
            color: 'lightgrey'
        },
        curveType: 'function',
        chartArea: { top: 25, bottom: 50, left: 50, right: 25 },
        hAxis: {
            gridlines: { count: 0 },
            minorGridlines: { count: 0 },
        },
        vAxis: {
            gridlines: { count: 0 },
            minorGridlines: { count: 0 },
            viewWindowMode: 'maximized',
        },
        // TODO: Figure out how to change annotation color based on the series above...
        annotations: {
            textStyle: {
              bold: true,
              italic: true,
              // The color of the text.
              color: 'green',
            }
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