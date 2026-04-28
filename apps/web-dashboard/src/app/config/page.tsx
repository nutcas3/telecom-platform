"use client";

import { useEffect, useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Switch } from "@/components/ui/switch";
import { 
  Settings, 
  Save, 
  RefreshCw, 
  Check, 
  X, 
  AlertTriangle,
  Search,
  Plus,
  Edit,
  Trash2,
  Eye,
  EyeOff
} from "lucide-react";
import { api, ConfigEntry, ConfigValidationResult } from "@/lib/api";

export default function ConfigPage() {
  const [configs, setConfigs] = useState<ConfigEntry[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [searchTerm, setSearchTerm] = useState("");
  const [showAddForm, setShowAddForm] = useState(false);
  const [editingConfig, setEditingConfig] = useState<ConfigEntry | null>(null);
  const [actionLoading, setActionLoading] = useState<string | null>(null);
  const [validationResult, setValidationResult] = useState<ConfigValidationResult | null>(null);
  const [showValidation, setShowValidation] = useState(false);
  
  // Form state
  const [formData, setFormData] = useState({
    key: "",
    value: "",
    type: "string",
    description: "",
    sensitive: false
  });

  useEffect(() => {
    loadConfigs();
  }, []);

  const loadConfigs = async () => {
    try {
      setLoading(true);
      const configsResponse = await api.config.get();
      setConfigs(Array.isArray(configsResponse) ? configsResponse : [configsResponse]);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load configuration");
    } finally {
      setLoading(false);
    }
  };

  const handleSave = async () => {
    try {
      setActionLoading('save');
      if (editingConfig) {
        // Update existing config
        await api.config.update(editingConfig.key, formData.value, formData.type);
      } else {
        // Create new config
        await api.config.update(formData.key, formData.value, formData.type);
      }
      
      // Reset form and reload data
      setFormData({ key: "", value: "", type: "string", description: "", sensitive: false });
      setEditingConfig(null);
      setShowAddForm(false);
      loadConfigs();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to save configuration");
    } finally {
      setActionLoading(null);
    }
  };

  const handleDelete = async (key: string) => {
    try {
      setActionLoading('delete');
      await api.config.delete(key);
      loadConfigs();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to delete configuration");
    } finally {
      setActionLoading(null);
    }
  };

  const handleValidate = async () => {
    try {
      setActionLoading('validate');
      const configObject = configs.reduce((acc, config) => {
        acc[config.key] = config.type === 'boolean' ? config.value === 'true' : 
                         config.type === 'number' ? parseFloat(config.value) : 
                         config.value;
        return acc;
      }, {} as Record<string, any>);
      
      const result = await api.config.validate(configObject);
      setValidationResult(result);
      setShowValidation(true);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to validate configuration");
    } finally {
      setActionLoading(null);
    }
  };

  const startEdit = (config: ConfigEntry) => {
    setEditingConfig(config);
    setFormData({
      key: config.key,
      value: config.value,
      type: config.type,
      description: config.description,
      sensitive: config.sensitive
    });
    setShowAddForm(true);
  };

  const cancelEdit = () => {
    setEditingConfig(null);
    setFormData({ key: "", value: "", type: "string", description: "", sensitive: false });
    setShowAddForm(false);
  };

  const getTypeColor = (type: string) => {
    switch (type) {
      case "string":
        return "bg-blue-100 text-blue-800";
      case "number":
        return "bg-green-100 text-green-800";
      case "boolean":
        return "bg-purple-100 text-purple-800";
      case "json":
        return "bg-orange-100 text-orange-800";
      default:
        return "bg-gray-100 text-gray-800";
    }
  };

  const formatValue = (value: string, type: string, sensitive: boolean) => {
    if (sensitive) {
      return "********";
    }
    if (type === "boolean") {
      return value === "true" ? "True" : "False";
    }
    if (type === "json") {
      try {
        return JSON.stringify(JSON.parse(value), null, 2);
      } catch {
        return value;
      }
    }
    return value;
  };

  const filteredConfigs = configs.filter(config =>
    config.key.toLowerCase().includes(searchTerm.toLowerCase()) ||
    config.description.toLowerCase().includes(searchTerm.toLowerCase())
  );

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <h1 className="text-3xl font-bold">Configuration Management</h1>
        <div className="flex gap-2">
          <Button
            onClick={handleValidate}
            disabled={actionLoading === 'validate'}
            variant="outline"
            className="flex items-center gap-2"
          >
            <RefreshCw className="h-4 w-4" />
            {actionLoading === 'validate' ? 'Validating...' : 'Validate Config'}
          </Button>
          <Button
            onClick={() => setShowAddForm(true)}
            className="flex items-center gap-2"
          >
            <Plus className="h-4 w-4" />
            Add Config
          </Button>
        </div>
      </div>

      {/* Validation Result */}
      {showValidation && validationResult && (
        <Card className={validationResult.valid ? "border-green-200 bg-green-50" : "border-red-200 bg-red-50"}>
          <CardContent className="p-4">
            <div className="flex items-center gap-2">
              {validationResult.valid ? (
                <Check className="h-5 w-5 text-green-600" />
              ) : (
                <X className="h-5 w-5 text-red-600" />
              )}
              <span className={validationResult.valid ? "text-green-800" : "text-red-800"}>
                {validationResult.valid ? "Configuration is valid" : "Configuration has errors"}
              </span>
            </div>
            {!validationResult.valid && validationResult.errors.length > 0 && (
              <ul className="mt-2 text-red-700 text-sm">
                {validationResult.errors.map((error, index) => (
                  <li key={index}>- {error}</li>
                ))}
              </ul>
            )}
          </CardContent>
        </Card>
      )}

      {/* Add/Edit Config Form */}
      {showAddForm && (
        <Card className="p-6">
          <CardHeader>
            <CardTitle>{editingConfig ? "Edit Configuration" : "Add New Configuration"}</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div>
              <Label htmlFor="key">Configuration Key</Label>
              <Input
                id="key"
                value={formData.key}
                onChange={(e) => setFormData({...formData, key: e.target.value})}
                placeholder="e.g., database.max_connections"
                disabled={!!editingConfig}
              />
            </div>
            <div>
              <Label htmlFor="value">Value</Label>
              {formData.type === "json" ? (
                <Textarea
                  id="value"
                  value={formData.value}
                  onChange={(e) => setFormData({...formData, value: e.target.value})}
                  placeholder="JSON value"
                  rows={4}
                />
              ) : (
                <Input
                  id="value"
                  value={formData.value}
                  onChange={(e) => setFormData({...formData, value: e.target.value})}
                  placeholder="Configuration value"
                  type={formData.sensitive ? "password" : "text"}
                />
              )}
            </div>
            <div>
              <Label htmlFor="type">Type</Label>
              <Select value={formData.type} onValueChange={(value) => setFormData({...formData, type: value})}>
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="string">String</SelectItem>
                  <SelectItem value="number">Number</SelectItem>
                  <SelectItem value="boolean">Boolean</SelectItem>
                  <SelectItem value="json">JSON</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div>
              <Label htmlFor="description">Description</Label>
              <Input
                id="description"
                value={formData.description}
                onChange={(e) => setFormData({...formData, description: e.target.value})}
                placeholder="Configuration description"
              />
            </div>
            <div className="flex items-center space-x-2">
              <Switch
                id="sensitive"
                checked={formData.sensitive}
                onCheckedChange={(checked) => setFormData({...formData, sensitive: checked})}
              />
              <Label htmlFor="sensitive">Sensitive (will be masked)</Label>
            </div>
            <div className="flex gap-2">
              <Button 
                onClick={handleSave}
                disabled={actionLoading === 'save' || !formData.key || !formData.value}
                className="flex items-center gap-2"
              >
                <Save className="h-4 w-4" />
                {actionLoading === 'save' ? 'Saving...' : 'Save Config'}
              </Button>
              <Button variant="outline" onClick={cancelEdit}>
                Cancel
              </Button>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Search */}
      <div className="relative">
        <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 h-4 w-4" />
        <Input
          placeholder="Search configurations..."
          value={searchTerm}
          onChange={(e) => setSearchTerm(e.target.value)}
          className="pl-10"
        />
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
            <div className="text-center text-gray-500">Loading configuration...</div>
          </CardContent>
        </Card>
      ) : (
        <div className="space-y-4">
          {filteredConfigs.map((config) => (
            <Card key={config.key}>
              <CardContent className="p-6">
                <div className="flex justify-between items-start">
                  <div className="flex-1">
                    <div className="flex items-center gap-2 mb-2">
                      <Settings className="h-5 w-5 text-gray-600" />
                      <h3 className="font-semibold">{config.key}</h3>
                      <Badge className={getTypeColor(config.type)}>
                        {config.type}
                      </Badge>
                      {config.sensitive && (
                        <Badge variant="outline">
                          <EyeOff className="h-3 w-3 mr-1" />
                          Sensitive
                        </Badge>
                      )}
                    </div>
                    <p className="text-gray-600 text-sm mb-2">{config.description}</p>
                    <div className="bg-gray-50 p-3 rounded-md">
                      <pre className="text-sm text-gray-800 whitespace-pre-wrap">
                        {formatValue(config.value, config.type, config.sensitive)}
                      </pre>
                    </div>
                    <div className="text-xs text-gray-500 mt-2">
                      Last updated: {new Date(config.updated_at).toLocaleString()}
                    </div>
                  </div>
                  <div className="flex gap-2 ml-4">
                    <Button variant="outline" size="sm" onClick={() => startEdit(config)}>
                      <Edit className="h-4 w-4" />
                    </Button>
                    <Button 
                      variant="outline" 
                      size="sm" 
                      onClick={() => handleDelete(config.key)}
                      disabled={actionLoading === 'delete'}
                    >
                      <Trash2 className="h-4 w-4" />
                    </Button>
                  </div>
                </div>
              </CardContent>
            </Card>
          ))}
          {filteredConfigs.length === 0 && (
            <Card>
              <CardContent className="p-8">
                <div className="text-center text-gray-500">
                  <Settings className="h-12 w-12 mx-auto mb-4 text-gray-300" />
                  <p>No configurations found</p>
                </div>
              </CardContent>
            </Card>
          )}
        </div>
      )}
    </div>
  );
}
