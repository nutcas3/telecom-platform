"use client";

import { useEffect, useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { 
  DollarSign, 
  FileText, 
  CreditCard, 
  Calendar, 
  Download,
  Eye,
  Plus,
  Search,
  Filter
} from "lucide-react";
import { api, Invoice, Payment } from "@/lib/api";

export default function BillingPage() {
  const [invoices, setInvoices] = useState<Invoice[]>([]);
  const [payments, setPayments] = useState<Payment[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [activeTab, setActiveTab] = useState<"invoices" | "payments">("invoices");
  
  // Filters
  const [searchTerm, setSearchTerm] = useState("");
  const [statusFilter, setStatusFilter] = useState<string>("");
  const [page, setPage] = useState(1);
  const [pageSize] = useState(20);
  
  // Generate invoice form
  const [showGenerateForm, setShowGenerateForm] = useState(false);
  const [generateForm, setGenerateForm] = useState({
    subscriberId: "",
    month: "",
    year: new Date().getFullYear().toString()
  });
  const [actionLoading, setActionLoading] = useState<string | null>(null);

  useEffect(() => {
    loadData();
  }, [activeTab, statusFilter, page]);

  const loadData = async () => {
    try {
      setLoading(true);
      if (activeTab === "invoices") {
        const invoicesResponse = await api.billing.invoices(undefined, statusFilter, page, pageSize);
        setInvoices(invoicesResponse.data);
      } else {
        const paymentsResponse = await api.billing.payments(undefined, statusFilter, page, pageSize);
        setPayments(paymentsResponse.data);
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load billing data");
    } finally {
      setLoading(false);
    }
  };

  const handleGenerateInvoice = async () => {
    try {
      setActionLoading('generate');
      await api.billing.generateInvoice(
        parseInt(generateForm.subscriberId),
        generateForm.month,
        parseInt(generateForm.year)
      );
      
      // Reset form and reload data
      setGenerateForm({ subscriberId: "", month: "", year: new Date().getFullYear().toString() });
      setShowGenerateForm(false);
      loadData();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to generate invoice");
    } finally {
      setActionLoading(null);
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case "completed":
      case "paid":
        return "bg-green-100 text-green-800";
      case "pending":
        return "bg-yellow-100 text-yellow-800";
      case "failed":
        return "bg-red-100 text-red-800";
      case "refunded":
        return "bg-blue-100 text-blue-800";
      default:
        return "bg-gray-100 text-gray-800";
    }
  };

  const formatCurrency = (amount: number, currency: string) => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: currency || 'USD'
    }).format(amount);
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString();
  };

  const filteredInvoices = invoices.filter(invoice =>
    invoice.id.toString().includes(searchTerm) ||
    invoice.status.toLowerCase().includes(searchTerm.toLowerCase())
  );

  const filteredPayments = payments.filter(payment =>
    payment.transaction_id.toLowerCase().includes(searchTerm.toLowerCase()) ||
    payment.status.toLowerCase().includes(searchTerm.toLowerCase())
  );

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <h1 className="text-3xl font-bold">Billing & Invoicing</h1>
        <div className="flex gap-2">
          <Button
            onClick={() => setShowGenerateForm(true)}
            className="flex items-center gap-2"
          >
            <Plus className="h-4 w-4" />
            Generate Invoice
          </Button>
        </div>
      </div>

      {/* Generate Invoice Modal */}
      {showGenerateForm && (
        <Card className="p-6">
          <CardHeader>
            <CardTitle>Generate New Invoice</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div>
              <Label htmlFor="subscriberId">Subscriber ID</Label>
              <Input
                id="subscriberId"
                type="number"
                value={generateForm.subscriberId}
                onChange={(e) => setGenerateForm({...generateForm, subscriberId: e.target.value})}
                placeholder="Enter subscriber ID"
              />
            </div>
            <div>
              <Label htmlFor="month">Month</Label>
              <Select value={generateForm.month} onValueChange={(value) => setGenerateForm({...generateForm, month: value})}>
                <SelectTrigger>
                  <SelectValue placeholder="Select month" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="1">January</SelectItem>
                  <SelectItem value="2">February</SelectItem>
                  <SelectItem value="3">March</SelectItem>
                  <SelectItem value="4">April</SelectItem>
                  <SelectItem value="5">May</SelectItem>
                  <SelectItem value="6">June</SelectItem>
                  <SelectItem value="7">July</SelectItem>
                  <SelectItem value="8">August</SelectItem>
                  <SelectItem value="9">September</SelectItem>
                  <SelectItem value="10">October</SelectItem>
                  <SelectItem value="11">November</SelectItem>
                  <SelectItem value="12">December</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div>
              <Label htmlFor="year">Year</Label>
              <Input
                id="year"
                type="number"
                value={generateForm.year}
                onChange={(e) => setGenerateForm({...generateForm, year: e.target.value})}
                placeholder="Enter year"
              />
            </div>
            <div className="flex gap-2">
              <Button 
                onClick={handleGenerateInvoice}
                disabled={actionLoading === 'generate'}
                className="flex items-center gap-2"
              >
                <FileText className="h-4 w-4" />
                {actionLoading === 'generate' ? 'Generating...' : 'Generate Invoice'}
              </Button>
              <Button 
                variant="outline" 
                onClick={() => setShowGenerateForm(false)}
              >
                Cancel
              </Button>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Tabs */}
      <div className="flex space-x-1 bg-gray-100 p-1 rounded-lg">
        <button
          onClick={() => setActiveTab("invoices")}
          className={`flex-1 py-2 px-4 rounded-md transition-colors ${
            activeTab === "invoices"
              ? "bg-white text-gray-900 shadow-sm"
              : "text-gray-600 hover:text-gray-900"
          }`}
        >
          <FileText className="h-4 w-4 inline mr-2" />
          Invoices
        </button>
        <button
          onClick={() => setActiveTab("payments")}
          className={`flex-1 py-2 px-4 rounded-md transition-colors ${
            activeTab === "payments"
              ? "bg-white text-gray-900 shadow-sm"
              : "text-gray-600 hover:text-gray-900"
          }`}
        >
          <CreditCard className="h-4 w-4 inline mr-2" />
          Payments
        </button>
      </div>

      {/* Filters */}
      <div className="flex flex-col sm:flex-row gap-4">
        <div className="flex-1">
          <div className="relative">
            <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 h-4 w-4" />
            <Input
              placeholder="Search..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="pl-10"
            />
          </div>
        </div>
        <Select value={statusFilter} onValueChange={setStatusFilter}>
          <SelectTrigger className="w-48">
            <Filter className="h-4 w-4 mr-2" />
            <SelectValue placeholder="Filter by status" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="">All Statuses</SelectItem>
            <SelectItem value="pending">Pending</SelectItem>
            <SelectItem value="completed">Completed</SelectItem>
            <SelectItem value="failed">Failed</SelectItem>
            <SelectItem value="refunded">Refunded</SelectItem>
          </SelectContent>
        </Select>
      </div>

      {error && (
        <Card className="border-red-200 bg-red-50">
          <CardContent className="p-4">
            <p className="text-red-800">{error}</p>
          </CardContent>
        </Card>
      )}

      {loading ? (
        <Card>
          <CardContent className="p-8">
            <div className="text-center text-gray-500">Loading billing data...</div>
          </CardContent>
        </Card>
      ) : (
        <>
          {activeTab === "invoices" ? (
            <div className="space-y-4">
              {filteredInvoices.map((invoice) => (
                <Card key={invoice.id}>
                  <CardContent className="p-6">
                    <div className="flex justify-between items-start">
                      <div>
                        <div className="flex items-center gap-2 mb-2">
                          <FileText className="h-5 w-5 text-blue-600" />
                          <h3 className="font-semibold">Invoice #{invoice.id}</h3>
                          <Badge className={getStatusColor(invoice.status)}>
                            {invoice.status}
                          </Badge>
                        </div>
                        <div className="grid grid-cols-1 md:grid-cols-3 gap-4 text-sm text-gray-600">
                          <div>
                            <Calendar className="h-4 w-4 inline mr-1" />
                            Due: {formatDate(invoice.due_date)}
                          </div>
                          <div>
                            <DollarSign className="h-4 w-4 inline mr-1" />
                            Amount: {formatCurrency(invoice.amount, invoice.currency)}
                          </div>
                          <div>
                            Created: {formatDate(invoice.created_at)}
                          </div>
                        </div>
                      </div>
                      <div className="flex gap-2">
                        <Button variant="outline" size="sm">
                          <Eye className="h-4 w-4 mr-1" />
                          View
                        </Button>
                        <Button variant="outline" size="sm">
                          <Download className="h-4 w-4 mr-1" />
                          Download
                        </Button>
                      </div>
                    </div>
                  </CardContent>
                </Card>
              ))}
              {filteredInvoices.length === 0 && (
                <Card>
                  <CardContent className="p-8">
                    <div className="text-center text-gray-500">
                      <FileText className="h-12 w-12 mx-auto mb-4 text-gray-300" />
                      <p>No invoices found</p>
                    </div>
                  </CardContent>
                </Card>
              )}
            </div>
          ) : (
            <div className="space-y-4">
              {filteredPayments.map((payment) => (
                <Card key={payment.id}>
                  <CardContent className="p-6">
                    <div className="flex justify-between items-start">
                      <div>
                        <div className="flex items-center gap-2 mb-2">
                          <CreditCard className="h-5 w-5 text-green-600" />
                          <h3 className="font-semibold">Payment #{payment.id}</h3>
                          <Badge className={getStatusColor(payment.status)}>
                            {payment.status}
                          </Badge>
                        </div>
                        <div className="grid grid-cols-1 md:grid-cols-3 gap-4 text-sm text-gray-600">
                          <div>
                            <DollarSign className="h-4 w-4 inline mr-1" />
                            Amount: {formatCurrency(payment.amount, payment.currency)}
                          </div>
                          <div>
                            Method: {payment.method}
                          </div>
                          <div>
                            <Calendar className="h-4 w-4 inline mr-1" />
                            {formatDate(payment.created_at)}
                          </div>
                        </div>
                        <div className="mt-2 text-sm text-gray-500">
                          Transaction ID: {payment.transaction_id}
                        </div>
                      </div>
                      <div className="flex gap-2">
                        <Button variant="outline" size="sm">
                          <Eye className="h-4 w-4 mr-1" />
                          View
                        </Button>
                      </div>
                    </div>
                  </CardContent>
                </Card>
              ))}
              {filteredPayments.length === 0 && (
                <Card>
                  <CardContent className="p-8">
                    <div className="text-center text-gray-500">
                      <CreditCard className="h-12 w-12 mx-auto mb-4 text-gray-300" />
                      <p>No payments found</p>
                    </div>
                  </CardContent>
                </Card>
              )}
            </div>
          )}
        </>
      )}
    </div>
  );
}
