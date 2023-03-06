import { Chart } from 'react-google-charts';


export default function StockLineChart({ chartData }) {

    function getFilteredChartData() {
        var data = []
        if ((chartData != null) && (chartData.sp500 != null)) {
            // Put the sorted values in an array, and add a column header.
            data = chartData.sp500.date.map((k, i) => [k, chartData.sp500.close[i]])
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
        chartArea: { top: 0, bottom: 0, left: 25, right: 25 }
    }
    var data = getFilteredChartData()

    return (
        <Chart
            chartType="LineChart"
            width="80%"
            height="600px"
            data={data}
            options={chartOptions}
        />
    );
}