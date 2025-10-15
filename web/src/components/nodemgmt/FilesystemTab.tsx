import { HardDrive } from "lucide-react";

interface FilesystemTabProps {
    filesystemMounts?: {
        mountpoint: string;
        device: string;
        type: string;
        size: number;
        used: number;
    }[];
}

const FilesystemTab = ({ filesystemMounts }: FilesystemTabProps) => {
  return (
    <div className="space-y-6">
        <h3 className="text-lg text-slate-800">Filesystem Management</h3>
        <div className="space-y-4">
        {!filesystemMounts || filesystemMounts.length === 0 ? (
                <p className="text-sm text-slate-600">No filesystem mounts available.</p>
        ) : (
            <>
                <h4 className="text-md text-slate-800">Mount Points</h4>
                {filesystemMounts?.map((mount) => (
                    <div key={mount.mountpoint} className="border border-slate-200 rounded-lg p-4">
                        <div className="flex items-center justify-between mb-3">
                            <div className="flex items-center space-x-3">
                                <HardDrive className="w-5 h-5 text-purple-600" />
                                <div>
                                    <h5 className="text-slate-800">{mount.mountpoint}</h5>
                                    <p className="text-sm text-slate-600">{mount.device} ({mount.type})</p>
                                </div>
                            </div>
                            <div className="text-right">
                                <p className="text-sm text-slate-800">{mount.used} G / {mount.size} G</p>
                                <p className="text-sm text-slate-600">{mount.size - mount.used} G available</p>
                            </div>
                        </div>
                        <div className="w-full bg-slate-200 rounded-full h-2">
                            <div 
                                className={`h-2 rounded-full ${
                                    (mount.used / mount.size) * 100 > 80 ? 'bg-red-500' : 
                                    (mount.used / mount.size) * 100 > 60 ? 'bg-yellow-500' : 'bg-green-500'
                                }`}
                                style={{ width: `${(mount.used / mount.size) * 100}%` }}
                            ></div>
                        </div>
                        <p className="text-xs text-slate-600 mt-1">{(mount.used / mount.size) * 100}% used</p>
                    </div>
                ))}
            </>
        )
        }
        </div>
    </div>
  )
}

export default FilesystemTab
