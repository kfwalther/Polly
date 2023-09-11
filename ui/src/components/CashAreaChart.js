import { Chart } from 'react-google-charts';

// An area chart to display percentage of portfolio that is cash.
export default function CashAreaChart({ chartDataSeries, chartTitle, startDate }) {

    // Define the options for this area chart.
    var chartOptions = {
        legend: 'none',
        backgroundColor: 'transparent',
        title: chartTitle,
        titleTextStyle: {
            color: 'lightgrey'
        },
        colors: ['aqua', 'darkgrey'],
        crosshair: { orientation: 'vertical', trigger: 'focus', color: 'lightgrey' },
        curveType: 'function',
        chartArea: { top: 25, bottom: 50, left: 65, right: 25, backgroundColor: '#0F0F0F' },
        hAxis: {
            format: 'MMM y',
            gridlines: { count: 5, color: 'transparent' },
            minorGridlines: { count: 0 },
        },
        isStacked: 'relative',
        vAxis: {
            format: 'percent',
            gridlines: { count: 0 },
            minorGridlines: { count: 0 },
            viewWindowMode: 'maximized',
        },
        responsive: true,
    }

    // Define the height of the chart range filter so we can pad for it.
    var filterHeight = 50

    // Use given start date, if one was provided.
    var plotStartDate = startDate == null ? chartDataSeries[1][0] : startDate

    return (
        // Apply bottom margin around this div container to account for the height of chart range filter.
        <div className="txnchart-container" style={{ marginBottom: filterHeight + 25 }}>
            <Chart
                chartType="AreaChart"
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
                                chartType: "AreaChart",
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
                                    // Set the range filter at some position.
                                    start: plotStartDate
                                },
                            },
                        },
                    },
                ]}
            />
        </div >
    );
}