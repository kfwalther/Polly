import { toUSD, toPercent } from './Helpers'
import './PortfolioSummary.css'

// Define the default function to export from this file.
export default function PortfolioSummary({ summaryData }) {

    // Format the gain/loss summary info with dollar amount, and percentage shown.
    function totalGainSummary() {
        let net = summaryData.totalMarketValue - summaryData.totalCostBasis;
        let plus = (net >= 0.0) ? "+" : "";
        return plus + toUSD(net) + " (" + toPercent(summaryData.percentageGain) + ")"
    }

    // Format the YTD total gain, and percentage gain.
    function ytdGainSummary() {
        let ytdPercentGain = summaryData.annualPerformance[new Date().getFullYear()]
        let net = summaryData.totalMarketValue - summaryData.totalMarketValue / (1 + ytdPercentGain / 100.0);
        let plus = (net >= 0.0) ? "+" : "";
        return plus + toUSD(net) + " (" + toPercent(ytdPercentGain) + ")"
    }
    
    // Calculate and format the daily gain summary metrics.
    function dailyGainSummary() {
        let dailyGainPercent = summaryData.dailyGain / (summaryData.totalMarketValue - summaryData.dailyGain) * 100;
        let plus = (summaryData.dailyGain >= 0.0) ? "+" : "";
        return plus + toUSD(summaryData.dailyGain) + " (" + toPercent(dailyGainPercent) + ")"
    }

    function renderHistoryTable() {
        return (
            <table className="historical-table">
                <thead>
                    <tr>
                        <th>Year</th>
                        {Object.entries(summaryData.annualPerformance).map(([year]) => (
                            <th key={year}>{year}</th>
                        ))}
                    </tr>
                </thead>
                <tbody>
                    <tr>
                        <td>Gain</td>
                        {Object.values(summaryData.annualPerformance).map((percentage, index) => (
                            <td key={index}>{toPercent(percentage)}</td>
                        ))}
                    </tr>
                </tbody>
            </table>
        )
    }

    return (
        <div className="summary-table-container">
            <div className="summary-table-div">
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
                            <td className="summary-deltas" style={{ color: summaryData.annualPerformance[new Date().getFullYear()] >= 0.0 ? 'green' : 'red' }}>
                                {ytdGainSummary()}
                            </td>
                            <td className="summary-cell" style={{ fontSize: '12px', color: 'grey' }}>YTD</td>
                        </tr>
                        <tr>
                            <td className="summary-deltas" style={{ color: summaryData.percentageGain >= 0.0 ? 'green' : 'red' }}>
                                {totalGainSummary()}
                            </td>
                            <td className="summary-cell" style={{ fontSize: '12px', color: 'grey' }}>ALL</td>
                        </tr>
                    </tbody>
                </table>
                <div className="summary-lastupdate" style={{ fontSize: '13px', color: 'grey' }}>
                    {"Last Updated: " + new Date(summaryData.lastUpdated).toISOString()}
                </div>
            </div>
            <div className="historical-table-div">
                {summaryData.annualPerformance ? renderHistoryTable() : null}
            </div>
        </div> 
    );
}

