package xyz.theditor.interlocktest

import android.content.res.AssetManager
import android.os.Bundle
import android.view.View
import android.widget.Button
import android.widget.LinearLayout
import android.widget.TextView
import androidx.activity.enableEdgeToEdge
import androidx.appcompat.app.AppCompatActivity
import androidx.core.view.ViewCompat
import androidx.core.view.WindowInsetsCompat

class MainActivity : AppCompatActivity() {

    companion object {
        init {
            System.loadLibrary("glue")
        }
    }

    external fun startRuntime(assetManager: AssetManager)

    fun createView(type: String, text: String?): View {
        return when (type.lowercase()) {
            "text" -> TextView(this).apply { this.text = text }
            "button" -> Button(this).apply { this.text = text }
            else -> View(this)
        }
    }

    fun createViews(items: Array<Pair<String, String?>>): List<View> {
        return items.map { (type, text) -> createView(type, text) }
    }

    fun enqueueCreateViews(types: Array<String>, texts: Array<String>) {
        runOnUiThread {
            val items = types.zip(texts).map { (type, text) -> type to text }
            val createdViews = createViews(items.toTypedArray())
            val container = findViewById<LinearLayout>(R.id.renderContainer)
            container.removeAllViews()
            createdViews.forEach { view ->
                container.addView(view)
            }
        }
    }

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        startRuntime(assets)
        enableEdgeToEdge()
        setContentView(R.layout.activity_main)
        ViewCompat.setOnApplyWindowInsetsListener(findViewById(R.id.main)) { v, insets ->
            val systemBars = insets.getInsets(WindowInsetsCompat.Type.systemBars())
            v.setPadding(systemBars.left, systemBars.top, systemBars.right, systemBars.bottom)
            insets
        }
    }
}