import { Chart } from 'react-google-charts';


export default function StockBarChart({ chartData, chartOptions }) {

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