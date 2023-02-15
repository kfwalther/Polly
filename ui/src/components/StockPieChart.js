import { Chart } from 'react-google-charts';


export default function StockPieChart({ chartData, chartOptions, filterOptions }) {

    function getFilteredChartData() {
        // Filter for only non-zero securities.
        let filtered = chartData.filter(s => (s.marketValue >= 0.0));
        // Check if we should be filtering for only stocks too.
        if (filterOptions) {
            filtered = filtered.filter(s => s.securityType === "Stock");
        }
        // Sort the stocks by current market value.
        let sorted = filtered.sort((a, b) => b.marketValue - a.marketValue);
        // Put the sorted values in an array, and add a column header.
        let marketValData = sorted.map(s => [s.ticker, s.marketValue])
        marketValData.unshift(['Ticker', 'Market Value'])
        return marketValData
    }

    return (
        <>
            <Chart
                chartType="PieChart"
                data={getFilteredChartData()}
                options={chartOptions}
                width={"100%"}
                height={"750px"}
            />
        </>
    );
}

