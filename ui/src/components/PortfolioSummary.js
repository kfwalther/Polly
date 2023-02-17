import { toUSD, toPercent } from './Helpers'

// Define the default function to export from this file.
export default function PortfolioSummary({ summaryData }) {

    // Format the gain/loss summary info with dollar amount, and percentage shown.
    function totalGainSummary() {
        let net = summaryData.totalMarketValue - summaryData.totalCostBasis;
        let plus = (net >= 0.0) ? "+" : "";
        return plus + toUSD(net) + " (" + toPercent(summaryData.percentageGain) + ")"
    }

    function dailyGainSummary() {
        let dailyGainPercent = summaryData.dailyGain / (summaryData.totalMarketValue - summaryData.dailyGain) * 100;
        let plus = (summaryData.dailyGain >= 0.0) ? "+" : "";
        return plus + toUSD(summaryData.dailyGain) + " (" + toPercent(dailyGainPercent) + ")"
    }

    return (
        <table className="summary-table">
            <tbody>
                <tr>
                    <td className="summary-cell" style={{ fontWeight: 'bold' }}>My Portfolio</td>
                </tr>
                <tr>
                    <td className="summary-cell" style={{ fontSize: '36px', fontWeight: 'bold' }}>{toUSD(summaryData.totalMarketValue)}</td>
                </tr>
                <tr>
                    <td className="summary-deltas" style={{ color: summaryData.dailyGain >= 0.0 ? 'green' : 'red' }}>
                        {dailyGainSummary()}
                    </td>
                    <td className="summary-cell" style={{ fontSize: '12px', color: 'grey' }}>TODAY</td>
                </tr>
                <tr>
                    <td className="summary-deltas" style={{ color: summaryData.percentageGain >= 0.0 ? 'green' : 'red' }}>
                        {totalGainSummary()}
                    </td>
                    <td className="summary-cell" style={{ fontSize: '12px', color: 'grey' }}>ALL</td>
                </tr>
            </tbody>
        </table >
    );
}

