import { Chart } from 'react-google-charts';


export default function StockLineChart({ chartData }) {

    function getFilteredChartData() {
        var data = []
        if ((chartData != null) && (chartData.sp500 != null)) {
            // Put the sorted values in an array, and add a column header.
            data = chartData.sp500.date.map((k, i) => [new Date(k), chartData.sp500.close[i]])
            data.unshift(['Date', 'Price'])
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