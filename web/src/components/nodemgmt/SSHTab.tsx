import React from "react";
import { Copy, Key, Plus, Trash2 } from "lucide-react";
import { toast } from "sonner";
import { handleAPIError } from "@/utils/errorHandler";
import { useNodeSSHKeys } from "@/services/hooks/useNodeSSHKeys";
import { sshKeysAPI } from "@/services/api/sshKeys";

interface SSHTabProps {
  nodeId: number;
  systemUsers?: string[];
}

function truncateKey(key: string) {
  const parts = key.trim().split(/\s+/);
  if (parts.length < 2) return key;
  const data = parts[1];
  if (data.length <= 16) return key;
  return `${parts[0]} ${data.slice(0, 10)}...${data.slice(-6)}${parts.slice(2).length ? " " + parts.slice(2).join(" ") : ""}`;
}

const SSHTab = ({ nodeId, systemUsers }: SSHTabProps) => {
  const { data: keys, loading, error, refetch } = useNodeSSHKeys(nodeId, { pollingInterval: 30000 });

  const usersForSelect = React.useMemo(() => {
    const all = [...(systemUsers ?? []), "root"].filter(Boolean);
    return Array.from(new Set(all)).sort();
  }, [systemUsers]);

  const [selectedMode, setSelectedMode] = React.useState<"existing" | "custom">(
    usersForSelect.length > 0 ? "existing" : "custom"
  );
  const [selectedUser, setSelectedUser] = React.useState<string>(usersForSelect[0] ?? "root");
  const [customUser, setCustomUser] = React.useState<string>("");

  const [publicKeyText, setPublicKeyText] = React.useState<string>("");
  const [comment, setComment] = React.useState<string>("");
  const [submitting, setSubmitting] = React.useState<boolean>(false);

  const [generated, setGenerated] = React.useState<null | { publicKey: string; privateKeyPem: string }>(null);

  const username = selectedMode === "custom" ? customUser.trim() : selectedUser;

  const handleAddKey = async () => {
    if (!username) {
      toast.error("Please choose a username.");
      return;
    }
    if (!publicKeyText.trim()) {
      toast.error("Please paste a public key.");
      return;
    }

    setSubmitting(true);
    try {
      await sshKeysAPI.createForNode(nodeId, {
        username,
        public_key: publicKeyText,
        comment: comment || undefined,
      });
      toast.success("SSH key saved.");
      setPublicKeyText("");
      setComment("");
      await refetch();
    } catch (e) {
      toast.error(handleAPIError(e, "create SSH key"));
    } finally {
      setSubmitting(false);
    }
  };

  const handleGenerateKey = async () => {
    if (!username) {
      toast.error("Please choose a username.");
      return;
    }

    setSubmitting(true);
    try {
      const res = await sshKeysAPI.generateForNode(nodeId, {
        username,
        comment: comment || undefined,
      });
      setGenerated({ publicKey: res.public_key, privateKeyPem: res.private_key_pem });
      toast.success("Generated keypair. Copy the private key now.");
      await refetch();
    } catch (e) {
      toast.error(handleAPIError(e, "generate SSH key"));
    } finally {
      setSubmitting(false);
    }
  };

  const handleDelete = async (keyId: number) => {
    if (!confirm("Delete this SSH key?")) return;
    setSubmitting(true);
    try {
      await sshKeysAPI.deleteForNode(nodeId, keyId);
      toast.success("SSH key deleted.");
      await refetch();
    } catch (e) {
      toast.error(handleAPIError(e, "delete SSH key"));
    } finally {
      setSubmitting(false);
    }
  };

  const list = keys ?? [];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h3 className="text-lg text-slate-800">SSH Key Management</h3>
      </div>

      <div className="border border-slate-200 rounded-lg p-4 space-y-4">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div>
            <p className="text-sm text-slate-600 mb-1">User on node</p>
            <div className="flex items-center space-x-2">
              <select
                className="border border-slate-300 rounded-lg px-3 py-2 text-sm w-full"
                value={selectedMode === "custom" ? "__custom__" : selectedUser}
                onChange={(e) => {
                  const v = e.target.value;
                  if (v === "__custom__") {
                    setSelectedMode("custom");
                  } else {
                    setSelectedMode("existing");
                    setSelectedUser(v);
                  }
                }}
              >
              {usersForSelect.map((u) => (
                <option key={u} value={u}>
                  {u}
                </option>
              ))}
                <option value="__custom__">Custom… (create if missing)</option>
              </select>
            </div>
            {selectedMode === "custom" && (
              <input
                className="mt-2 border border-slate-300 rounded-lg px-3 py-2 text-sm w-full"
                placeholder="e.g. alukacs"
                value={customUser}
                onChange={(e) => setCustomUser(e.target.value)}
              />
            )}
            <p className="mt-2 text-xs text-slate-500">
              Agent creates the user if it doesn&apos;t exist, then adds the key to <code className="font-mono">authorized_keys</code>.
            </p>
          </div>

          <div>
            <p className="text-sm text-slate-600 mb-1">Comment (optional)</p>
            <input
              className="border border-slate-300 rounded-lg px-3 py-2 text-sm w-full"
              placeholder="e.g. laptop-2026"
              value={comment}
              onChange={(e) => setComment(e.target.value)}
            />
          </div>
        </div>

        <div>
          <p className="text-sm text-slate-600 mb-1">Public key</p>
          <textarea
            className="border border-slate-300 rounded-lg px-3 py-2 text-sm w-full font-mono min-h-[100px]"
            placeholder="ssh-ed25519 AAAAC3... user@host"
            value={publicKeyText}
            onChange={(e) => setPublicKeyText(e.target.value)}
          />
        </div>

        <div className="flex items-center space-x-2">
          <button
            disabled={submitting}
            onClick={handleAddKey}
            className="px-4 py-2 bg-blue-600 hover:bg-blue-700 disabled:opacity-60 text-white rounded-lg text-sm transition-colors flex items-center space-x-2"
          >
            <Plus className="w-4 h-4" />
            <span>Add Key</span>
          </button>
          <button
            disabled={submitting}
            onClick={handleGenerateKey}
            className="px-4 py-2 bg-slate-800 hover:bg-slate-900 disabled:opacity-60 text-white rounded-lg text-sm transition-colors flex items-center space-x-2"
          >
            <Key className="w-4 h-4" />
            <span>Generate RSA Keypair</span>
          </button>
        </div>
      </div>

      {generated && (
        <div className="border border-yellow-200 bg-yellow-50 rounded-lg p-4 space-y-3">
          <p className="text-sm text-slate-800">
            Private key (copy now — it&apos;s not stored on the API):
          </p>
          <div className="flex items-center space-x-2">
            <button
              className="px-3 py-1 text-blue-700 hover:bg-blue-100 border border-blue-300 rounded text-sm transition-colors flex items-center space-x-2"
              onClick={() => {
                navigator.clipboard.writeText(generated.privateKeyPem);
                toast.success("Copied private key.");
              }}
            >
              <Copy className="w-4 h-4" />
              <span>Copy</span>
            </button>
            <button
              className="px-3 py-1 text-slate-700 hover:bg-slate-100 border border-slate-300 rounded text-sm transition-colors"
              onClick={() => setGenerated(null)}
            >
              Dismiss
            </button>
          </div>
          <textarea
            className="w-full font-mono text-xs bg-white border border-yellow-200 rounded-lg p-3 min-h-[160px]"
            value={generated.privateKeyPem}
            readOnly
          />
          <p className="text-xs text-slate-600">
            Public key stored for agent sync:
            <code className="ml-2 text-xs bg-white border border-yellow-200 px-2 py-1 rounded font-mono">{truncateKey(generated.publicKey)}</code>
          </p>
        </div>
      )}

      <div className="space-y-3">
        <div className="flex items-center justify-between">
          <h4 className="text-slate-800">Stored keys</h4>
          <p className="text-xs text-slate-500">Auto-updates every 30 seconds.</p>
        </div>

        {error ? (
          <div className="text-sm text-red-600">
            {error.message}{" "}
            <button className="ml-2 text-blue-600 hover:text-blue-800" onClick={() => refetch()}>
              Retry
            </button>
          </div>
        ) : loading && !keys ? (
          <p className="text-sm text-slate-600">Loading SSH keys…</p>
        ) : list.length === 0 ? (
          <p className="text-sm text-slate-600">No SSH keys yet.</p>
        ) : (
          <div className="space-y-2">
            {list.map((k) => (
              <div key={k.id} className="border border-slate-200 rounded-lg p-3 flex items-start justify-between">
                <div className="min-w-0">
                  <div className="flex items-center space-x-2">
                    <span className="text-sm text-slate-800 font-medium">{k.username}</span>
                    <span className="text-xs text-slate-500">{new Date(k.created_at).toLocaleString()}</span>
                  </div>
                  <div className="mt-1 flex items-center space-x-2">
                    <code className="text-xs text-slate-700 bg-slate-100 px-2 py-1 rounded font-mono truncate">
                      {truncateKey(k.public_key)}
                    </code>
                    <button
                      className="p-1 text-slate-600 hover:text-slate-800"
                      onClick={() => {
                        navigator.clipboard.writeText(k.public_key);
                        toast.success("Copied public key.");
                      }}
                    >
                      <Copy className="w-4 h-4" />
                    </button>
                  </div>
                </div>
                <button
                  disabled={submitting}
                  onClick={() => handleDelete(k.id)}
                  className="ml-3 px-3 py-1 text-red-600 hover:bg-red-50 border border-red-600 rounded text-sm transition-colors flex items-center space-x-2 disabled:opacity-60"
                >
                  <Trash2 className="w-4 h-4" />
                  <span>Delete</span>
                </button>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
};

export default SSHTab;
