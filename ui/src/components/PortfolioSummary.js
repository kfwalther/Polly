import toUSD from './Helpers'

export default function PortfolioSummary({ summaryData }) {

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
        let data = sorted.map(s => [s.ticker, s[displayDataset]]);
        data.unshift(['Ticker', 'Displayed Dataset']);
        return data;
    }

    function gainLossSummary() {
        let net = summaryData.totalMarketValue - summaryData.totalCostBasis;
        let plus = (net >= 0.0) ? "+" : "";
        return plus + toUSD(net) + " (" + summaryData.percentageGain.toFixed(2) + "%)"
    }

    return (
        <>
            <h3>My Portfolio</h3>
            <h3 className="summary-totalmarketvalue">{toUSD(summaryData.totalMarketValue)}</h3>
            <div className="summary-deltas" style={{ color: this.state.portfolioSummary.percentageGain >= 0.0 ? "green" : "red" }}>
                    {gainLossSummary()}
            </div>
            <h3 className="summary-deltas">{gainLoss()}</h3>
        </>
    );
}

