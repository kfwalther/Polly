import { Chart } from 'react-google-charts';


export default function StockPieChart({ chartData, chartOptions, displayDataset, filterOptions, title, titleDesc }) {

    function getFilteredChartData() {
        // Filter for only non-zero securities.
        let filtered = chartData.filter(s => (s[displayDataset] > 0.0));
        // Check if we should be filtering for only stocks too.
        if (filterOptions) {
            filtered = filtered.filter(s => s.securityType === "Stock");
        }
        // Sort the stocks by the dataset being displayed.
        let sorted = filtered.sort((a, b) => b[displayDataset] - a[displayDataset]);
        // Put the sorted values in an array, and add a column header.
        let data = sorted.map(s => [s.ticker, s[displayDataset]])
        data.unshift(['Ticker', 'Displayed Dataset'])
        return data
    }

    return (
        <>
            <div className="piechart-container">
                <Chart
                    chartType="PieChart"
                    data={getFilteredChartData()}
                    options={chartOptions}
                    width={"100%"}
                    height={"750px"}
                />
                <div className="piechart-underlaylabel">{titleDesc}</div>
                <div className="piechart-underlay">{title}</div>
            </div>
        </>
    );
}

