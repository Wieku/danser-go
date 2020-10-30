#version 330

in vec4 col_tint;
in vec3 tex_coord;
in float additive;
in float msdf;

uniform sampler2DArray tex;

out vec4 color;

float median(float r, float g, float b) {
	return max(min(r, g), min(max(r, g), b));
}

void main() {
    vec4 in_color = texture(tex, tex_coord);

	if (msdf > 0.5) {
		float sigDist = median(in_color.r, in_color.g, in_color.b);
		float w = fwidth(sigDist);
		float opacity = smoothstep(0.5 - w, 0.5 + w, sigDist);
		in_color = vec4(vec3(1.0), opacity);
	}

	color = in_color*col_tint;
	color.rgb *= color.a;
	color.a *= additive;
}