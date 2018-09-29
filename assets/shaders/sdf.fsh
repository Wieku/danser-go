#version 330

in vec4 col_tint;
in vec3 tex_coord;
in float additive;

uniform sampler2DArray tex;

out vec4 color;

const float smoothing = 1.0/8;

void main()
{
    float distance = texture(tex, tex_coord).a;
    float outlineFactor = smoothstep(0.5 - smoothing, 0.5 + smoothing, distance);
	color = vec4(1, 1, 1, outlineFactor)*col_tint;
	color.rgb *= color.a;
	if (additive == 1) {
	    color.a = 0;
	}
}