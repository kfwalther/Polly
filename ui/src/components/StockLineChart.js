import { Chart } from 'react-google-charts';


export default function StockLineChart({ chartDataSeries }) {

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

    // Define the height of the chart range filter so we can pad for it.
    var filterHeight = 50

    return (
        // Apply bottom margin around this div container to account for the height of chart range filter.
        <div className="txnchart-container" style={{ marginBottom: filterHeight + 25 }}>
            <Chart
                chartType="LineChart"
                width="100%"
                height="600px"
                data={chartDataSeries}
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