import { useState, useEffect } from "react";
import { TransactionsTable } from "./TransactionsTable";
import { toPercent } from "./Helpers";

// Fetch the transaction data from the server, and return/render the transactions page.
export default function TransactionsPage() {
    const [txnList, setTxnList] = useState([]);

    document.body.style.backgroundColor = "black"

    // Fetch the transaction data from the server.
    useEffect(() => {
        console.log('Fetching transaction data...')
        fetch("http://localhost:5000/transactions")
            .then(response => response.json())
            .then(resp => setTxnList(resp["transactions"]))
    }, []);

    return (
        <>
            {/* Display a rollup/summary of all our transactions. */}
            <table className="txn-summary-table">
                <th className="txn-summary-table-header">
                    Transaction Count
                </th>
                <th className="txn-summary-table-header">
                    Win Rate
                </th>
                <th className="txn-summary-table-header">
                    Beat Rate
                </th>
                <tbody>
                    <tr>
                        <td className="txn-summary-table-cell">{txnList.length}</td>
                        <td className="txn-summary-table-cell">{toPercent((txnList.filter(x => x.totalReturn > 0).length / txnList.length) * 100.0)}</td>
                        <td className="txn-summary-table-cell">{toPercent((txnList.filter(x => x.excessReturn > 0).length / txnList.length) * 100.0)}</td>
                    </tr>
                </tbody>
            </table>
            {/* Display all the transactions in a sortable table. */}
            <TransactionsTable
                txnData={txnList.filter(t => (t.action === "Buy" || t.action === "Sell"))}
            />
        </>
    );
}