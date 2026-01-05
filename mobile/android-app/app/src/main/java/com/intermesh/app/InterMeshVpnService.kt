package com.intermesh.app

import android.content.Intent
import android.net.ProxyInfo
import android.net.VpnService
import android.os.ParcelFileDescriptor
import android.util.Log

class InterMeshVpnService : VpnService() {

    private var vpnInterface: ParcelFileDescriptor? = null

    override fun onStartCommand(intent: Intent?, flags: Int, startId: Int): Int {
        val action = intent?.action
        if (action == ACTION_STOP) {
            stopVpn()
            return START_NOT_STICKY
        }

        startVpn()
        return START_STICKY
    }

    private fun startVpn() {
        if (vpnInterface != null) return

        try {
            val builder =
                    Builder()
                            .setSession("InterMesh VPN")
                            .addAddress("10.0.0.1", 24)
                            .addDnsServer("8.8.8.8")
                            .addRoute("0.0.0.0", 0)

            // Set HTTP proxy for the VPN interface (API 29+)
            // This tells Android to route all HTTP/HTTPS traffic through our local proxy
            if (android.os.Build.VERSION.SDK_INT >= android.os.Build.VERSION_CODES.Q) {
                builder.setHttpProxy(ProxyInfo.buildDirectProxy("127.0.0.1", 8080))
            }

            vpnInterface = builder.establish()
            Log.d(TAG, "VPN Interface established")
        } catch (e: Exception) {
            Log.e(TAG, "Failed to establish VPN: ${e.message}")
        }
    }

    private fun stopVpn() {
        try {
            vpnInterface?.close()
            vpnInterface = null
            Log.d(TAG, "VPN Interface closed")
        } catch (e: Exception) {
            Log.e(TAG, "Failed to close VPN: ${e.message}")
        }
        stopSelf()
    }

    override fun onDestroy() {
        super.onDestroy()
        stopVpn()
    }

    companion object {
        private const val TAG = "InterMeshVpnService"
        const val ACTION_STOP = "com.intermesh.app.STOP_VPN"
    }
}
