import { toUSD } from './Helpers'

export default function PortfolioSummary({ summaryData }) {

    function gainLossSummary() {
        let net = summaryData.totalMarketValue - summaryData.totalCostBasis;
        let plus = (net >= 0.0) ? "+" : "";
        return plus + toUSD(net) + " (" + summaryData.percentageGain.toFixed(2) + "%)"
    }

    return (
        <>
            <h3>My Portfolio</h3>
            <h3 className="summary-totalmarketvalue">{toUSD(summaryData.totalMarketValue)}</h3>
            <div className="summary-deltas" style={{ color: summaryData.percentageGain >= 0.0 ? "green" : "red" }}>
                {gainLossSummary()}
            </div>
        </>
    );
}

