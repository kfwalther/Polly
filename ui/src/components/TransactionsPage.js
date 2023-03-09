import { useState, useEffect } from "react";
import { TransactionsTable } from "./TransactionsTable";
import { toPercent } from "./Helpers";
import StockLineChart from "./StockLineChart";

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

    return (
        <>
            {/* Display a rollup/summary of all our transactions. */}
            <table className="txn-summary-table">
                <th className="txn-summary-table-header">Transaction Count</th>
                <th className="txn-summary-table-header">Win Rate</th>
                <th className="txn-summary-table-header">Beat Rate</th>
                <tbody>
                    <tr>
                        <td className="txn-summary-table-cell">{txnList.length}</td>
                        <td className="txn-summary-table-cell">{toPercent((txnList.filter(x => x.totalReturn > 0).length / txnList.length) * 100.0)}</td>
                        <td className="txn-summary-table-cell">{toPercent((txnList.filter(x => x.excessReturn > 0).length / txnList.length) * 100.0)}</td>
                    </tr>
                </tbody>
            </table>
            {/* Add a line chart to superimpose our trades on the S&P500 performance. */}
            <StockLineChart
                chartData={sp500}
            />
            <h3 className="header-centered">Transactions List</h3>
            {/* Display all the transactions in a sortable table. */}
            <TransactionsTable
                txnData={txnList.filter(t => (t.action === "Buy" || t.action === "Sell"))}
            />
        </>
    );
}