#version 330

in vec3 in_position;
in vec3 in_mid;
in vec2 in_tex_coord;
in float in_index;
in float hue;

uniform mat4 proj;
uniform float scale;
uniform float points;
uniform float endScale;
uniform float hueshift;

out vec2 tex_coord;
out vec4 color_pass;
out float index;

vec4 hsv2rgb(float h,float s,float v) {
	const float eps=1e-3;
	vec4 result=vec4(0.0, 0.0, 0.0, 1.0);
	if(s<=0.0)result.r=result.g=result.b=v;
	else {
		float hi=floor(h/60.0);
		float f=(h/60.0)-hi;
		float m=v*(1.0-s);
		float n=v*(1.0-s*f);
		float k=v*(1.0-s*(1.0-f));
		if(hi<=0.0+eps) {
			result.r=v;
			result.g=k;
			result.b=m;
		} else if(hi<=1.0+eps) {
			result.r=n;
			result.g=v;
			result.b=m;
		} else if(hi<=2.0+eps) {
			result.r=m;
			result.g=v;
			result.b=k;
		} else if(hi<=3.0+eps) {
			result.r=m;
			result.g=n;
			result.b=v;
		} else if(hi<=4.0+eps) {
			result.r=k;
			result.g=m;
			result.b=v;
		} else if(hi<=5.0+eps) {
			result.r=v;
			result.g=m;
			result.b=n;
		}
	}
	return result;
}

void main() {
    gl_Position = proj * vec4((in_position - in_mid) * scale * (endScale + (1.0 - endScale) * (points-1-in_index) / points) + in_mid, 1);
    tex_coord = in_tex_coord;
	index = in_index;
	color_pass = vec4(hsv2rgb(fract(hue+hueshift)*360, 1., 1.).rgb, 1.0);
}