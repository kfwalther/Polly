import { useState, useEffect } from "react";
import { TransactionsTable } from "./TransactionsTable";
import { toPercent } from "./Helpers";
import StockLineChart from "./StockLineChart";

// Helper function to calculate the win rate for our trades (buys/sells).
function getWinRate(txns) {
    if (txns.length > 0) {
        return toPercent((txns.filter(x => x.totalReturn > 0).length / txns.length) * 100.0)
    } else {
        return toPercent(0.0)
    }
}

// Helper function to calculate the market beat rate for our trades (buys/sells).
function getBeatRate(txns) {
    if (txns.length > 0) {
        return toPercent((txns.filter(x => x.excessReturn > 0).length / txns.length) * 100.0)
    } else {
        return toPercent(0.0)
    }
}

// Fetch the transaction data from the server, and return/render the transactions page.
export default function TransactionsPage() {
    const [txnList, setTxnList] = useState([]);
    const [sp500, setSp500] = useState({});

    document.body.style.backgroundColor = "black"

    // Fetch the transaction data from the server.
    useEffect(() => {
        console.log('Fetching transaction data...')
        fetch("http://localhost:5000/transactions")
            .then(response => response.json())
            .then(resp => setTxnList(resp["transactions"]))
        fetch("http://localhost:5000/sp500")
            .then(response => response.json())
            .then(resp => setSp500(resp))
    }, []);

    var buySellList = []
    // Filter txns for only buy/sell actions.
    if (txnList != null) {
        buySellList = txnList.filter(t => (t.action === "Buy" || t.action === "Sell"))
    }

    return (
        <>
            {/* Display a rollup/summary of all our transactions. */}
            <table className="txn-summary-table">
                <th className="txn-summary-table-header">Transaction Count</th>
                <th className="txn-summary-table-header">Win Rate</th>
                <th className="txn-summary-table-header">Beat Rate</th>
                <tbody>
                    <tr>
                        <td className="txn-summary-table-cell">{buySellList.length}</td>
                        <td className="txn-summary-table-cell">{getWinRate(buySellList)}</td>
                        <td className="txn-summary-table-cell">{getBeatRate(buySellList)}</td>
                    </tr>
                </tbody>
            </table>
            {/* Add a line chart to superimpose our trades on the S&P500 performance. */}
            <StockLineChart
                chartData={sp500}
                txnData={buySellList}
            />
            <h3 className="header-centered">Transactions List</h3>
            {/* Display all the transactions in a sortable table. */}
            <TransactionsTable
                txnData={buySellList}
            />
        </>
    );
}