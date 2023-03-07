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
        title: "S&P500 Performance",
        curveType: "function",
        chartArea: { top: 25, bottom: 50, left: 50, right: 25 },
        responsive: true,
    }
    var data = getFilteredChartData()

    return (
        // <div className="txnchart-container">
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
                            chartType: "LineChart",
                            chartOptions: {
                                backgroundColor: "black",
                                chartArea: { width: "95%", height: "90%" },
                                hAxis: { baselineColor: "none" },
                            },
                        },
                    },
                    controlPosition: "bottom",
                    controlWrapperParams: {
                        state: {
                            range: {
                                start: new Date(2020, 1, 1),
                                end: new Date(2023, 2, 28),
                            },
                        },
                    },
                },
            ]}
        />
        // </div>
    );
}