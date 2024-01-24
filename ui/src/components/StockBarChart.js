import { Chart } from 'react-google-charts';


export default function StockBarChart({ chartData }) {

    function getFilteredChartData() {
        // Filter for only non-zero equities.
        let filtered = chartData.filter(s => (s.marketValue > 0.0));
        // Sort the stocks by the dataset being displayed.
        let sorted = filtered.sort((a, b) => b.unrealizedGain - a.unrealizedGain);
        // Put the sorted values in an array, and add a column header.
        let data = sorted.map(s => [s.ticker, s.unrealizedGain, s.unrealizedGain >= 0 ? 'green' : 'red'])
        data.unshift(['Ticker', 'Unrealized Gain', { role: 'style' }])
        return data
    }

    var data = getFilteredChartData()
    // Define the options for the bar chart.
    var chartOptions = {
        backgroundColor: 'black',
        legend: { position: 'none' },
        chartArea: { top: 25, bottom: 50, left: 40, right: 40 },
        vAxis: { format: 'short', textStyle: { fontSize: 12, bold: true, color: 'grey' } },
        hAxis: { showTextEvery: 1, maxAlternation: 1, slantedText: true, slantedTextAngle: 45, textStyle: { fontSize: 12, bold: true, color: 'grey' } },
        bar: { groupWidth: '40%' }
    }
    return (
        <div className="barchart-container">
            <Chart
                chartType="ColumnChart"
                width={(data.length - 1) * 50}
                height="400px"
                data={data}
                options={chartOptions}
            />
        </div>
    );
}